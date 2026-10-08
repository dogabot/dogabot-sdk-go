// Fetch a ticker with query params (GET /api/v1/ticker).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	dogabot "github.com/dogabot/dogabot-sdk-go"
)

func main() {
	if os.Getenv("DOGABOT_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "Set DOGABOT_API_KEY=dbk_live_...")
		os.Exit(1)
	}
	exchange := envOr("DOGABOT_EXCHANGE", "hyperliquid")
	symbol := envOr("DOGABOT_SYMBOL", "BTC")
	c := dogabot.New(os.Getenv("DOGABOT_API_KEY"))
	q := url.Values{}
	q.Set("exchange", exchange)
	q.Set("symbol", symbol)
	ticker, err := c.GetTicker(context.Background(), dogabot.WithQuery(q))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(ticker)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
