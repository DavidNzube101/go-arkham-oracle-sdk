package go_arkham_oracle_sdk

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/crypto/sha3"
)

// SignedPriceData holds the data returned from the oracle API.
type SignedPriceData struct {
	Price     uint64
	Timestamp int64
	Signature []byte
}

// internal struct for JSON unmarshalling of oracle API response
type priceResponse struct {
	Price     string `json:"price"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

// internal struct for JSON unmarshalling of CoinGecko-like price data
type CoinGeckoPriceResponse map[string]struct {
	Usd float64 `json:"usd"`
}

// Client is a client for the Arkham Oracle API.
type Client struct {
	BaseURL string
	DataSourceURL string // Optional: URL for an alternative data source
}

// NewClient creates a new oracle client.
// baseURL is the full base URL of the price API endpoint (e.g., "https://arkham-dvpn.vercel.app/api/price").
// dataSourceURL is optional. If provided, FetchSignedPrice will use this URL to get raw price data.
func NewClient(baseURL string, dataSourceURL ...string) *Client {
	client := &Client{BaseURL: baseURL}
	if len(dataSourceURL) > 0 && dataSourceURL[0] != "" {
		client.DataSourceURL = dataSourceURL[0]
	}
	return client
}

// FetchSignedPrice fetches signed price data from the oracle.
// The trustedKey is optional. If provided, it will be sent as a query parameter.
func (c *Client) FetchSignedPrice(token string, trustedKey ...string) (*SignedPriceData, error) {
	// Determine the URL to fetch the raw price data from
	priceSourceURL := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", token)
	if c.DataSourceURL != "" {
		// Assuming the custom data source uses similar query parameters
		priceSourceURL = fmt.Sprintf("%s?ids=%s&vs_currencies=usd", c.DataSourceURL, token)
	}

	// Fetch raw price data from the determined source
	priceResp, err := http.Get(priceSourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch raw price data from %s: %w", priceSourceURL, err)
	}
	defer priceResp.Body.Close()

	if priceResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("raw price data source returned non-200 status: %s", priceResp.Status)
	}

	var rawPriceData CoinGeckoPriceResponse
	if err := json.NewDecoder(priceResp.Body).Decode(&rawPriceData); err != nil {
		return nil, fmt.Errorf("failed to decode raw price data: %w", err)
	}

	priceFloat := rawPriceData[token].Usd
	if priceFloat == 0 { // CoinGecko returns 0 if token not found
		return nil, fmt.Errorf("price for token '%s' not found in data source response", token)
	}

	// Convert price to micro-units (6 decimals) as a uint64
	priceU64 := uint64(priceFloat * 1_000_000)
	timestampI64 := time.Now().Unix()

	// --------------------------------------------------------------------
	// Now, call the oracle API to get the signed data
	// --------------------------------------------------------------------
	params := url.Values{}
	params.Add("token", token)
	params.Add("price", strconv.FormatUint(priceU64, 10))
	params.Add("timestamp", strconv.FormatInt(timestampI64, 10))
	if len(trustedKey) > 0 && trustedKey[0] != "" {
		params.Add("trustedClientKey", trustedKey[0])
	}
	reqURL := fmt.Sprintf("%s?%s", c.BaseURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call oracle API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return nil, fmt.Errorf("oracle API returned non-200 status: %s - %s", resp.Status, errorBody.Error)
	}

	var rawResp priceResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode oracle API response: %w", err)
	}

	// Parse the string data into the correct types
	price, err := strconv.ParseUint(rawResp.Price, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from API: %w", err)
	}
	timestamp, err := strconv.ParseInt(rawResp.Timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp from API: %w", err)
	}
	signature, err := hex.DecodeString(rawResp.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature from API: %w", err)
	}

	return &SignedPriceData{
		Price:     price,
		Timestamp: timestamp,
		Signature: signature,
	}, nil
}

// CreateOracleMessageHash reconstructs the 32-byte Keccak-256 hash from the price and timestamp.
// This is the message that was signed by the oracle.
func (d *SignedPriceData) CreateOracleMessageHash() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, d.Price); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, d.Timestamp); err != nil {
		return nil, err
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(buf.Bytes())
	return hasher.Sum(nil), nil
}
