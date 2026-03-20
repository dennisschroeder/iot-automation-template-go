package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/dennisschroeder/iot-automation-template-go/internal/mqtt"
	"github.com/dennisschroeder/iot-automation-template-go/internal/nats"
	"github.com/dennisschroeder/iot-automation-template-go/internal/service"
)

var (
	RootCmd = &cobra.Command{
		Use:   "iot-automation-template-go",
		Short: "A template for event-driven IoT automations in Go",
		Run:   runRoot,
	}

	rootArgs struct {
		logLevel string
		mqttHost string
		mqttPort int
		clientID string
		natsURL  string
		service  service.Config
	}
)

func init() {
	f := RootCmd.Flags()

	// Generic flags
	f.StringVar(&rootArgs.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	// MQTT flags
	f.StringVar(&rootArgs.mqttHost, "mqtt-host", "mosquitto.mqtt.svc.cluster.local", "MQTT broker host")
	f.IntVar(&rootArgs.mqttPort, "mqtt-port", 1883, "MQTT broker port")
	f.StringVar(&rootArgs.clientID, "mqtt-client-id", "iot-automation-template", "MQTT Client ID")

	// NATS flags
	f.StringVar(&rootArgs.natsURL, "nats-url", "nats://nats.event-bus.svc.cluster.local:4222", "NATS server URL")
}

func runRoot(cmd *cobra.Command, args []string) {
	// Setup structured logging (slog)
	var level slog.Level
	if err := level.UnmarshalText([]byte(rootArgs.logLevel)); err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	slog.Info("Starting iot-automation-template-go...")

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Prepare NATS Client
	natsClient, err := nats.NewClient(rootArgs.natsURL)
	if err != nil {
		slog.Error("Failed to initialize NATS client", "error", err)
		os.Exit(1)
	}
	defer natsClient.Close()

	// 2. Prepare MQTT Client
	mqttBroker := fmt.Sprintf("tcp://%s:%d", rootArgs.mqttHost, rootArgs.mqttPort)
	mqttClient, err := mqtt.NewClient(mqttBroker, rootArgs.clientID)
	if err != nil {
		slog.Error("Failed to initialize MQTT client", "error", err)
		os.Exit(1)
	}
	defer mqttClient.Close()

	// 3. Prepare and run Main Service (Automation Logic)
	svc := service.New(rootArgs.service, logger, natsClient, mqttClient)

	// Listen for OS signals to trigger graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal, terminating...")
		cancel()
	}()

	// Block and run
	if err := svc.Run(ctx); err != nil {
		slog.Error("Service exited with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Service stopped cleanly")
}
