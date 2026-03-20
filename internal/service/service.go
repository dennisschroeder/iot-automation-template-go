package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/dennisschroeder/iot-app-template-go/internal/mqtt"
	"github.com/dennisschroeder/iot-app-template-go/internal/source"
)

// Config represents service-level settings
type Config struct {
	PollInterval time.Duration
}

// Service contains the main business logic
type Service struct {
	config Config
	logger *slog.Logger
	mqtt   *mqtt.Client
	source *source.Client
}

// New creates a new Service instance
func New(cfg Config, logger *slog.Logger, srcClient *source.Client, mqttClient *mqtt.Client) *Service {
	return &Service{
		config: cfg,
		logger: logger,
		mqtt:   mqttClient,
		source: srcClient,
	}
}

// Run starts the main loop of the application
func (s *Service) Run(ctx context.Context) error {
	s.logger.Info("Starting service loop", "poll_interval", s.config.PollInterval)

	// Publish Initial MQTT Auto-Discovery Payloads
	s.mqtt.PublishDiscovery("temperature_1", "Room Temperature", "measurement", "temperature", "°C")
	s.mqtt.PublishDiscovery("pressure_1", "Room Pressure", "measurement", "pressure", "bar")
	s.mqtt.PublishDiscovery("humidity_1", "Room Humidity", "measurement", "humidity", "%")

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	// Perform an initial poll
	if err := s.pollAndPublish(ctx); err != nil {
		s.logger.Error("Initial poll failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping service loop")
			return nil
		case <-ticker.C:
			if err := s.pollAndPublish(ctx); err != nil {
				s.logger.Error("Error during polling cycle", "error", err)
			}
		}
	}
}

func (s *Service) pollAndPublish(ctx context.Context) error {
	s.logger.Debug("Polling source data...")

	// 1. Fetch data from source (Modbus, REST, etc.)
	data, err := s.source.FetchData(ctx)
	if err != nil {
		return err
	}

	// 2. Publish data to MQTT for Home Assistant
	for key, value := range data {
		s.mqtt.PublishState(key, value)
		s.logger.Debug("Published sensor data", "sensor", key, "value", value)
	}

	return nil
}
