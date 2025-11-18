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
	oracleClient := oraclesdk.NewClient("https://arkham-dvpn.vercel.app/api/price")

	// 2. Fetch the signed price data

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

	// 3. Recreate the message hash that was signed
	// This is required for the Ed25519Program instruction
	messageHash, err := signedData.CreateOracleMessageHash()
	if err != nil {
		log.Fatalf("Failed to create message hash: %v", err)
	}

	fmt.Printf("\nRecreated Message Hash: %x\n", messageHash)
	
	// The signature and messageHash can now be used to build a transaction.
}
```

```