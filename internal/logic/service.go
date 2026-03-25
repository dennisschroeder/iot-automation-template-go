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

const (
	TargetPIR   = "everything_presence_one_office_pir"
	TargetLight = "of_desk_sublight"
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
	slog.Info("Starting PIR-to-Light Automation...", "pir", TargetPIR, "light", TargetLight)

	// Subscribe to IoT events on NATS
	_, err := s.nats.Subscribe("iot.events.>", func(msg *natsgo.Msg) {
		slog.Debug("NATS message received", "subject", msg.Subject, "size", len(msg.Data))
		var envelope iotv1.EventEnvelope
		if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
			slog.Debug("Skipping non-envelope message or unmarshal error", "subject", msg.Subject)
			return
		}

		// Check if it's a PresenceEvent for our target PIR
		if presence := envelope.GetPresence(); presence != nil {
			slog.Debug("Processing presence event", "entity_id", presence.EntityId, "target", TargetPIR)
			if presence.EntityId == TargetPIR {
				slog.Info("Presence match found", "state", presence.State.String())
				s.handlePresence(presence)
			}
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

func (s *Service) handlePresence(presence *iotv1.PresenceEvent) {
	var stateStr string
	if presence.State == iotv1.BinaryState_BINARY_STATE_ON {
		stateStr = "ON"
	} else {
		stateStr = "OFF"
	}

	slog.Info("Switching light", "light", TargetLight, "state", stateStr)

	// Direct MQTT action for the "Durchstich"
	// Subject pattern: cmnd/<light_id>/POWER
	topic := fmt.Sprintf("cmnd/%s/POWER", TargetLight)
	if err := s.mqtt.Publish(topic, stateStr); err != nil {
		slog.Error("Failed to publish light command", "topic", topic, "error", err)
	}
}
