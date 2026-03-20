package service

import (
	"context"
	"log/slog"

	"github.com/dennisschroeder/iot-automation-template-go/internal/mqtt"
	"github.com/dennisschroeder/iot-automation-template-go/internal/nats"
)

// Config represents service-level settings
type Config struct {
}

// Service contains the main business logic
type Service struct {
	config Config
	logger *slog.Logger
	nats   *nats.Client
	mqtt   *mqtt.Client
}

// New creates a new Service instance
func New(cfg Config, logger *slog.Logger, natsClient *nats.Client, mqttClient *mqtt.Client) *Service {
	return &Service{
		config: cfg,
		logger: logger,
		nats:   natsClient,
		mqtt:   mqttClient,
	}
}

// Run starts the main loop of the application
func (s *Service) Run(ctx context.Context) error {
	s.logger.Info("Starting automation service loop")

	// Example: Register NATS watchers here
	// manager := watcher.NewManager(s.nats.JS())
	// manager.Register(&MyWatcher{...})

	<-ctx.Done()
	s.logger.Info("Context cancelled, stopping service loop")
	return nil
}
