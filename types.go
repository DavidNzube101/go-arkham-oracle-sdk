package go_arkham_oracle_sdk

// SignedPriceData holds the data returned from the oracle API.
// The struct tags ensure it can be easily marshalled to JSON with string-encoded numbers.
type SignedPriceData struct {
	Price     uint64 `json:"price,string"`
	Timestamp int64  `json:"timestamp,string"`
	Signature []byte `json:"signature"`
}

// CoinGeckoPriceResponse is an internal struct for JSON unmarshalling of CoinGecko-like price data.
type CoinGeckoPriceResponse map[string]struct {
	Usd float64 `json:"usd"`
}
