package go_arkham_oracle_sdk

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/sha3"
)

// OracleHandlerOptions configures the oracle handler.
type OracleHandlerOptions struct {
	// The 64-byte Ed25519 private key used for signing price data.
	OraclePrivateKey ed25519.PrivateKey
	// Optional. An array of strings to use as API keys for authorization.
	// If this slice is nil or empty, the endpoint will be public.
	TrustedClientKeys []string
	// Optional. A URL for an alternative data source to fetch prices from.
	// If empty, CoinGecko will be used by default.
	DataSourceURL string
}

// NewOracleHandler creates a new http.HandlerFunc that acts as a verifiable oracle.
func NewOracleHandler(options OracleHandlerOptions) (http.HandlerFunc, error) {
	if len(options.OraclePrivateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(options.OraclePrivateKey))
	}

	// Create a map for quick lookup of trusted keys
	trustedKeysMap := make(map[string]bool)
	for _, key := range options.TrustedClientKeys {
		trustedKeysMap[key] = true
	}

	// This is the actual handler function that will be returned
	handler := func(w http.ResponseWriter, r *http.Request) {
		// 1. Security Validation (Optional)
		if len(trustedKeysMap) > 0 {
			clientKey := r.URL.Query().Get("trustedClientKey")
			if !trustedKeysMap[clientKey] {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, `{"error":"Token parameter is required"}`, http.StatusBadRequest)
			return
		}

		// 2. Fetch Price from Data Source
		dataSource := "https://api.coingecko.com/api/v3/simple/price"
		if options.DataSourceURL != "" {
			dataSource = options.DataSourceURL
		}
		priceURL := fmt.Sprintf("%s?ids=%s&vs_currencies=usd", dataSource, token)

		priceResp, err := http.Get(priceURL)
		if err != nil {
			log.Printf("Error fetching from data source: %v", err)
			http.Error(w, `{"error":"Failed to fetch price data"}`, http.StatusInternalServerError)
			return
		}
		defer priceResp.Body.Close()

		var priceData CoinGeckoPriceResponse
		if err := json.NewDecoder(priceResp.Body).Decode(&priceData); err != nil {
			log.Printf("Error decoding price data: %v", err)
			http.Error(w, `{"error":"Failed to decode price data"}`, http.StatusInternalServerError)
			return
		}

		priceFloat := priceData[token].Usd
		if priceFloat == 0 {
			http.Error(w, fmt.Sprintf(`{"error":"Price for token '%s' not found"}`, token), http.StatusNotFound)
			return
		}

		// 3. Prepare Data for Signing
		priceU64 := uint64(priceFloat * 1_000_000)
		timestampI64 := time.Now().Unix()

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, priceU64)
		binary.Write(buf, binary.LittleEndian, timestampI64)
		
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(buf.Bytes())
		messageHash := hasher.Sum(nil)

		// 4. Sign the Message Hash
		signature := ed25519.Sign(options.OraclePrivateKey, messageHash)

		// 5. Return the Data
		responsePayload := SignedPriceData{
			Price:     priceU64,
			Timestamp: timestampI64,
			Signature: signature,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responsePayload)
	}

	return handler, nil
}
