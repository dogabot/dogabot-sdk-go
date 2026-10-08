// Place a tiny paper market order (Idempotency-Key required).
// Dry-run unless CONFIRM_PLACE=1. trading_mode stays paper.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	dogabot "github.com/dogabot/dogabot-sdk-go"
)

func main() {
	if os.Getenv("DOGABOT_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "Set DOGABOT_API_KEY=dbk_live_...")
		os.Exit(1)
	}

	body := map[string]any{
		"exchange":       envOr("DOGABOT_EXCHANGE", "hyperliquid"),
		"symbol":         envOr("DOGABOT_SYMBOL", "BTC"),
		"side":           "buy",
		"quantity":       0.001,
		"order_type":     "market",
		"trading_mode":   "paper",
		"broadcast_mode": "personal",
	}
	idem := os.Getenv("DOGABOT_IDEMPOTENCY_KEY")
	if idem == "" {
		var b [8]byte
		_, _ = rand.Read(b[:])
		idem = "example-paper-" + hex.EncodeToString(b[:])
	}

	if os.Getenv("CONFIRM_PLACE") != "1" {
		fmt.Println("Dry-run only. Would POST placeorder with:")
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"body": body, "idempotency_key": idem})
		fmt.Println("Re-run with CONFIRM_PLACE=1 to submit (still paper).")
		return
	}

	c := dogabot.New(os.Getenv("DOGABOT_API_KEY"))
	out, err := c.PostTerminalPlaceorder(context.Background(), body, dogabot.WithIdempotencyKey(idem))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
