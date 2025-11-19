# Arkham Oracle SDK (Go)

The `go-arkham-oracle-sdk` provides a robust client and server for interacting with a custom verifiable off-chain oracle. This SDK simplifies the process of fetching and serving cryptographically signed price data, ensuring authenticity and integrity for Go applications that need to deliver this data to critical contexts like smart contracts.

## Installation

```bash
go get github.com/DavidNzube101/go-arkham-oracle-sdk
```

## Usage

### For Clients: Fetching Signed Prices

The client helps you fetch the signed price data from an oracle and reconstruct the message hash that was signed. This is necessary for building the `Ed25519Program` instruction required by compatible smart contracts.

```go
package main

import (
	"fmt"
	"log"
	"os"

	oraclesdk "github.com/DavidNzube101/go-arkham-oracle-sdk"
)

func main() {
	// 1. Initialize the client
	// By default, the oracle API uses CoinGecko as the raw price data source.
	oracleClient := oraclesdk.NewClient("https://BASE-URL/api/price")

	// --- Example 1: Fetching from a protected endpoint ---
	fmt.Println("--- Fetching from protected endpoint ---")
	token := "solana"
	trustedKey := os.Getenv("ORACLE_CLIENT_TRUSTED_KEY") // This key must be known by the oracle server
	if trustedKey == "" {
		log.Println("Warning: ORACLE_CLIENT_TRUSTED_KEY not set, assuming public endpoint.")
	}
	
signedData, err := oracleClient.FetchSignedPrice(token, trustedKey)
	if err != nil {
		log.Fatalf("Failed to fetch signed price: %v", err)
	}

	fmt.Printf("Successfully fetched data for %s:\n", token)
	fmt.Printf("- Price: %d\n", signedData.Price)
	fmt.Printf("- Timestamp: %d\n", signedData.Timestamp)
	fmt.Printf("- Signature: %x\n", signedData.Signature)

	// --- Example 2: Fetching from a public endpoint ---
	fmt.Println("\n--- Fetching from public endpoint ---")
	// Simply omit the trustedKey argument if the oracle server is configured to be public
	publicSignedData, err := oracleClient.FetchSignedPrice(token)
	if err != nil {
		log.Fatalf("Failed to fetch signed price from public endpoint: %v", err)
	}
	fmt.Printf("Successfully fetched data for %s:\n", token)
	fmt.Printf("- Price: %d\n", publicSignedData.Price)


	// --- Example 3: Using a custom data source for the oracle ---
	fmt.Println("\n--- Using a custom data source ---")
	// The oracle server will fetch raw price data from "https://my-custom-price-api.com/prices"
	// instead of CoinGecko.
	customOracleClient := oraclesdk.NewClient(
		"https://BASE-URL/api/price",
		"https://my-custom-price-api.com/prices", // Your custom data source URL
	)
	customSignedData, err := customOracleClient.FetchSignedPrice(token, trustedKey)
	if err != nil {
		log.Fatalf("Failed to fetch signed price from custom source: %v", err)
	}
	fmt.Printf("Successfully fetched data from custom source for %s:\n", token)
	fmt.Printf("- Price: %d\n", customSignedData.Price)


	// --- Recreate the message hash for a transaction ---
	// This is required for the Ed25519Program instruction in Solana smart contracts.
	messageHash, err := signedData.CreateOracleMessageHash()
	if err != nil {
		log.Fatalf("Failed to create message hash: %v", err)
	}

	fmt.Printf("\nRecreated Message Hash: %x\n", messageHash)
	
	// The signature and messageHash can now be used to build a Solana transaction.
}
```

### For Servers: Creating an Oracle API

You can easily create a verifiable oracle API endpoint using the `NewOracleHandler` function. This allows your Go application to act as a price oracle, signing data for clients.

```go
package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	oraclesdk "github.com/DavidNzube101/go-arkham-oracle-sdk"
)

func main() {
	// 1. Load your 64-byte Ed25519 private key
	// Example: ORACLE_PRIVATE_KEY="[56,120,167,...,215,160,71]"
	privateKeyStr := os.Getenv("ORACLE_PRIVATE_KEY")
	if privateKeyStr == "" {
		log.Fatal("ORACLE_PRIVATE_KEY environment variable is not set.")
	}
	var privateKeyBytes []byte
	if err := json.Unmarshal([]byte(privateKeyStr), &privateKeyBytes); err != nil {
		log.Fatalf("Failed to parse ORACLE_PRIVATE_KEY: %v", err)
	}
	oraclePrivateKey := ed25519.PrivateKey(privateKeyBytes)

	// 2. (Optional) Load trusted client keys
	// Example: TRUSTED_CLIENT_KEYS="key1,key2,key3"
	trustedClientKeysStr := os.Getenv("TRUSTED_CLIENT_KEYS")
	var trustedClientKeys []string
	if trustedClientKeysStr != "" {
		trustedClientKeys = strings.Split(trustedClientKeysStr, ",")
	}

	// 3. (Optional) Specify a custom data source URL
	dataSourceURL := os.Getenv("DATA_SOURCE_URL") // e.g., "https://my-custom-price-api.com/prices"

	// 4. Create the oracle handler
	oracleHandler, err := oraclesdk.NewOracleHandler(oraclesdk.OracleHandlerOptions{
		OraclePrivateKey:  oraclePrivateKey,
		TrustedClientKeys: trustedClientKeys, // Omit this field to make the endpoint public
		DataSourceURL:     dataSourceURL,     // Omit this field to use CoinGecko by default
	})
	if err != nil {
		log.Fatalf("Failed to create oracle handler: %v", err)
	}

	// 5. Set up the HTTP server
	http.Handle("/api/price", oracleHandler)

	port := ":8080"
	fmt.Printf("Go Oracle Server running on http://localhost%s/api/price\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
```

## Use Cases

This SDK is ideal for scenarios where:

*   **Decentralized Finance (DeFi):** Providing tamper-proof price feeds for lending protocols, stablecoins, or derivatives on Solana and other blockchains.
*   **Decentralized VPNs (dVPNs):** As implemented in the Arkham dVPN Protocol, for securely valuing staked assets or bandwidth payments.
*   **Gaming & NFTs:** Integrating real-world asset prices or dynamic game parameters into blockchain-based games or NFT marketplaces.
*   **Supply Chain & Logistics:** Verifying real-time commodity prices or sensor data on-chain.
*   **Any Smart Contract Requiring External Data:** When a smart contract needs to react to off-chain information, this oracle provides a cryptographically secure bridge.

## Author: (David Nzube)[https://github.com/DavidNzube101]