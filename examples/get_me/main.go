// Print the authenticated account (GET /api/v1/me).
package main

import (
	"context"
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
	c := dogabot.New(os.Getenv("DOGABOT_API_KEY"))
	me, err := c.GetMe(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(me)
}
