package nats

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewClient(url string) (*Client, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JetStream: %w", err)
	}

	slog.Info("NATS JetStream client initialized", "url", url)
	return &Client{nc: nc, js: js}, nil
}

func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.nc.Subscribe(subject, handler)
}

func (c *Client) Publish(subject string, data []byte) error {
	return c.nc.Publish(subject, data)
}

func (c *Client) JetStream() nats.JetStreamContext {
	return c.js
}

func (c *Client) Close() {
	c.nc.Close()
}
