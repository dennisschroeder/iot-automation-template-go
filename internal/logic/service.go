package logic

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dennisschroeder/iot-automation-template-go/internal/transport/mqtt"
	"github.com/dennisschroeder/iot-automation-template-go/internal/transport/nats"
	iotv1 "github.com/dennisschroeder/homelab-gitops/gen/go/iot/v1"
	natsgo "github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	nats *nats.Client
	mqtt *mqtt.Client
}

func NewService(n *nats.Client, m *mqtt.Client) *Service {
	return &Service{
		nats: n,
		mqtt: m,
	}
}

func (s *Service) Run(ctx context.Context) error {
	slog.Info("Starting IoT Automation logic...")

	// Example Watcher: Listen to typed Sensor Events
	_, err := s.nats.Subscribe("iot.sensors.>", func(msg *natsgo.Msg) {
		var event iotv1.SensorEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			slog.Error("Failed to unmarshal sensor event", "error", err)
			return
		}
		slog.Debug("Received typed sensor event", "id", event.Id, "entity", event.EntityId, "value", event.Value)
	})
	if err != nil {
		return err
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		slog.Info("Context cancelled, shutting down")
	case sig := <-stop:
		slog.Info("Signal received, shutting down", "signal", sig)
	}

	return nil
}
