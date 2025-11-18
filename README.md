# Arkham Oracle SDK (Go)

This package provides a client to interact with a custom Arkham-compatible oracle.

## Installation

```bash
go get github.com/arkham-org/go-arkham-oracle-sdk
```

## Usage

The client helps you fetch the signed price data from the oracle and reconstruct the message hash that was signed. This is necessary for building the `Ed25519Program` instruction required by compatible smart contracts.

```go
package main

import (
	"fmt"
	"log"

	oraclesdk "github.com/arkham-org/go-arkham-oracle-sdk"
)

func main() {
	// 1. Initialize the client
	// By default, it uses CoinGecko as the raw price data source.
	oracleClient := oraclesdk.NewClient("https://arkham-dvpn.vercel.app/api/price")

	// --- Example 1: Fetching from a protected endpoint ---
	fmt.Println("--- Fetching from protected endpoint ---")

token := "solana"
trustedKey := "your-trusted-client-key"
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
	// Simply omit the trustedKey argument
	publicSignedData, err := oracleClient.FetchSignedPrice(token)
	if err != nil {
		log.Fatalf("Failed to fetch signed price from public endpoint: %v", err)
	}
	fmt.Printf("Successfully fetched data for %s:\n", token)
	fmt.Printf("- Price: %d\n", publicSignedData.Price)


	// --- Example 3: Using a custom data source ---
	fmt.Println("\n--- Using a custom data source ---")
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
	// This is required for the Ed25519Program instruction
	messageHash, err := signedData.CreateOracleMessageHash()
	if err != nil {
		log.Fatalf("Failed to create message hash: %v", err)
	}

	fmt.Printf("\nRecreated Message Hash: %x\n", messageHash)
	
	// The signature and messageHash can now be used to build a transaction.
}
```
