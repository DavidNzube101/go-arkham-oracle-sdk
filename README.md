# Arkham Oracle SDK (Go)

The `go-arkham-oracle-sdk` provides a robust client for interacting with a custom verifiable off-chain oracle. This SDK simplifies the process of fetching cryptographically signed price data, ensuring authenticity and integrity for Go applications that need to deliver this data to critical contexts like smart contracts.

## Installation

```bash
go get github.com/arkham-org/go-arkham-oracle-sdk
```

## Usage

The client helps you fetch the signed price data from the oracle and reconstruct the message hash that was signed. This is necessary for building the `Ed25519Program` instruction required by compatible smart contracts.

### Basic Client Usage

```go
package main

import (
	"fmt"
	"log"

	oraclesdk "github.com/arkham-org/go-arkham-oracle-sdk"
)

func main() {
	// 1. Initialize the client
	// By default, the oracle API uses CoinGecko as the raw price data source.
	oracleClient := oraclesdk.NewClient("https://arkham-dvpn.vercel.app/api/price")

	// --- Example 1: Fetching from a protected endpoint ---
	fmt.Println("--- Fetching from protected endpoint ---")

token := "solana"
trustedKey := "your-trusted-client-key" // This key must be known by the oracle server
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
		"https://arkham-dvpn.vercel.app/api/price",
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

### Example: Serving Signed Data via a Go Webserver

This example demonstrates how a Go backend could fetch signed price data and then expose it via its own API endpoint.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	oraclesdk "github.com/arkham-org/go-arkham-oracle-sdk"
)

func main() {
	// Initialize the oracle client
	// Replace with your actual oracle API base URL
	oracleClient := oraclesdk.NewClient("https://arkham-dvpn.vercel.app/api/price")

	// (Optional) Get trusted client key from environment
	trustedClientKey := os.Getenv("ORACLE_CLIENT_TRUSTED_KEY")
	if trustedClientKey == "" {
		log.Println("Warning: ORACLE_CLIENT_TRUSTED_KEY not set, fetching from public endpoint.")
	}

	http.HandleFunc("/signed-price", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing 'token' query parameter", http.StatusBadRequest)
			return
		}

		// Fetch signed price data
		signedData, err := oracleClient.FetchSignedPrice(token, trustedClientKey)
		if err != nil {
			log.Printf("Error fetching signed price for %s: %v", token, err)
			http.Error(w, "Failed to fetch signed price", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(signedData)
	})

	port := ":8080"
	fmt.Printf("Go web server serving signed prices on http://localhost%s/signed-price\n", port)
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

```


## Author: (David Nzube)[https://github.com/DavidNzube101]
