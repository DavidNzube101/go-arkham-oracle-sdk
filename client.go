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

	"golang.org/x/crypto/sha3"
)

// SignedPriceData holds the data returned from the oracle API.
type SignedPriceData struct {
	Price     uint64
	Timestamp int64
	Signature []byte
}

// internal struct for JSON unmarshalling
type priceResponse struct {
	Price     string `json:"price"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

// Client is a client for the Arkham Oracle API.
type Client struct {
	BaseURL string
}

// NewClient creates a new oracle client.
// baseURL is the full base URL of the price API endpoint (e.g., "https://arkham-dvpn.vercel.app/api/price").
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// FetchSignedPrice fetches signed price data from the oracle.
func (c *Client) FetchSignedPrice(token, trustedKey string) (*SignedPriceData, error) {
	params := url.Values{}
	params.Add("token", token)
	params.Add("trustedClientKey", trustedKey)
	reqURL := fmt.Sprintf("%s?%s", c.BaseURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call price API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return nil, fmt.Errorf("price API returned non-200 status: %s - %s", resp.Status, errorBody.Error)
	}

	var rawResp priceResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode price API response: %w", err)
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
