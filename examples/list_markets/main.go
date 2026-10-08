// List markets with optional filters (GET /api/v1/markets).
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
	c := dogabot.New(os.Getenv("DOGABOT_API_KEY"))
	q := url.Values{}
	q.Set("limit", "5")
	q.Set("is_tradable", "true")
	markets, err := c.GetMarkets(context.Background(), dogabot.WithQuery(q))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(markets)
}
