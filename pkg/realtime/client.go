package realtime

import (
	"context"
	"fmt"

	"github.com/philippseith/signalr"
)

type receiver struct{}

type Client struct {
	address string
}

func NewClient(address string) *Client {
	return &Client{address: address}
}

func (c *Client) Invoke(ctx context.Context, method string, args ...any) error {
	clientCtx, clientCancel := context.WithCancel(ctx)
	defer clientCancel()

	conn, err := signalr.NewHTTPConnection(clientCtx, c.address)
	if err != nil {
		return fmt.Errorf("signalr connection: %w", err)
	}

	client, err := signalr.NewClient(clientCtx,
		signalr.WithConnection(conn),
		signalr.WithReceiver(&receiver{}))
	if err != nil {
		return fmt.Errorf("signalr client: %w", err)
	}

	client.Start()

	select {
	case result := <-client.Invoke(method, args...):
		return result.Error
	case <-ctx.Done():
		return ctx.Err()
	}
}
