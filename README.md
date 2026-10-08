# dogabot-sdk-go

Official Go client for the [dogabot public API](https://docs.dogabot.com/).

```bash
go get github.com/dogabot/dogabot-sdk-go
```

```go
package main

import (
	"context"
	"fmt"
	"os"

	dogabot "github.com/dogabot/dogabot-sdk-go"
)

func main() {
	c := dogabot.New(os.Getenv("DOGABOT_API_KEY"))
	me, err := c.GetMe(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", me)
}
```

Writes need `WithIdempotencyKey`:

```go
_, err := c.PostTerminalPlaceorder(ctx, map[string]any{
	"exchange": "hyperliquid", "symbol": "BTC", "side": "buy",
	"quantity": 0.001, "order_type": "market",
	"trading_mode": "paper", "broadcast_mode": "personal",
}, dogabot.WithIdempotencyKey("place-paper-btc-1"))
```

Runnable samples: [`examples/`](./examples/).

Docs: https://docs.dogabot.com/sdk/
