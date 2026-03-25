package logic

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	iotv1 "github.com/dennisschroeder/iot-schemas-proto/gen/go/iot/v1"
	"github.com/dennisschroeder/iot-automation-template-go/internal/transport/mqtt"
	"github.com/dennisschroeder/iot-automation-template-go/internal/transport/nats"
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
	slog.Info("Starting IoT Hierarchical Event Logger...")

	// Subscribe to ALL v1 events using the new hierarchical pattern
	// Pattern: iot.v1.events.<source>.<area>.<device_id>
	subject := "iot.v1.events.>"
	_, err := s.nats.Subscribe(subject, func(msg *natsgo.Msg) {
		var envelope iotv1.EventEnvelope
		if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
			slog.Warn("Failed to unmarshal v1 envelope", "subject", msg.Subject, "error", err)
			return
		}

		// Log detailed information based on payload type
		slog.Info("Event Received", 
			"subject", msg.Subject,
			"source", envelope.Source,
			"id", envelope.Id,
		)

		if presence := envelope.GetPresence(); presence != nil {
			slog.Info("├── Presence detected", 
				"entity", presence.EntityId, 
				"state", presence.State.String(),
			)
		} else if light := envelope.GetLight(); light != nil {
			slog.Info("├── Light state updated", 
				"entity", light.EntityId, 
				"state", light.State.String(),
				"brightness", fmt.Sprintf("%.2f", light.Brightness),
			)
		} else {
			slog.Info("└── Other payload type received")
		}
	})
	if err != nil {
		return fmt.Errorf("nats subscribe failed: %w", err)
	}

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
