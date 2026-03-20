package source

import (
	"context"
	"log/slog"
	"time"
)

// Config represents the connection to the source device (e.g. Modbus, API)
type Config struct {
	Address string
	Port    int
}

// Client represents the connection to the source data
type Client struct {
	config Config
	logger *slog.Logger
}

// NewClient creates a new mock/template source client
func NewClient(cfg Config, logger *slog.Logger) (*Client, error) {
	return &Client{
		config: cfg,
		logger: logger,
	}, nil
}

// Connect simulates connecting to a data source
func (c *Client) Connect(ctx context.Context) error {
	c.logger.Info("Connected to source device", "address", c.config.Address, "port", c.config.Port)
	return nil
}

func (c *Client) Disconnect() {
	c.logger.Info("Disconnected from source device")
}

// FetchData is a dummy method. Replace this with real Modbus/REST/TCP logic.
func (c *Client) FetchData(ctx context.Context) (map[string]float64, error) {
	// Simulate reading from a device
	time.Sleep(100 * time.Millisecond)

	// Return mock data
	return map[string]float64{
		"temperature_1": 21.5,
		"pressure_1":    1.2,
		"humidity_1":    45.0,
	}, nil
}
