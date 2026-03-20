package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/dennisschroeder/iot-app-template-go/internal/mqtt"
	"github.com/dennisschroeder/iot-app-template-go/internal/service"
	"github.com/dennisschroeder/iot-app-template-go/internal/source"
)

var (
	RootCmd = &cobra.Command{
		Use:   "iot-app-template-go",
		Short: "A template for stateless IoT bridges in Go",
		Run:   runRoot,
	}

	rootArgs struct {
		logLevel string
		mqtt     mqtt.Config
		source   source.Config
		service  service.Config
	}
)

func init() {
	f := RootCmd.Flags()

	// Generic flags
	f.StringVar(&rootArgs.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	// MQTT flags
	f.StringVar(&rootArgs.mqtt.Host, "mqtt-host", "mosquitto.mqtt.svc.cluster.local", "MQTT broker host")
	f.IntVar(&rootArgs.mqtt.Port, "mqtt-port", 1883, "MQTT broker port")
	f.StringVar(&rootArgs.mqtt.ClientID, "mqtt-client-id", "iot-app-template", "MQTT Client ID")

	// Source API / Modbus flags
	f.StringVar(&rootArgs.source.Address, "source-address", "192.168.1.100", "IP address of the source device")
	f.IntVar(&rootArgs.source.Port, "source-port", 80, "Port of the source device")

	// Service specific flags
	f.DurationVar(&rootArgs.service.PollInterval, "poll-interval", 30*time.Second, "Interval to poll the source device")
}

func runRoot(cmd *cobra.Command, args []string) {
	// Setup structured logging (slog)
	var level slog.Level
	if err := level.UnmarshalText([]byte(rootArgs.logLevel)); err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	slog.Info("Starting iot-app-template-go...")

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Prepare MQTT Client (Target)
	mqttClient, err := mqtt.NewClient(rootArgs.mqtt, logger)
	if err != nil {
		slog.Error("Failed to initialize MQTT client", "error", err)
		os.Exit(1)
	}
	if err := mqttClient.Connect(ctx); err != nil {
		slog.Error("Failed to connect to MQTT broker", "error", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect()

	// 2. Prepare Source Client (e.g., Modbus, REST)
	srcClient, err := source.NewClient(rootArgs.source, logger)
	if err != nil {
		slog.Error("Failed to initialize source client", "error", err)
		os.Exit(1)
	}
	if err := srcClient.Connect(ctx); err != nil {
		slog.Error("Failed to connect to source device", "error", err)
		os.Exit(1)
	}
	defer srcClient.Disconnect()

	// 3. Prepare and run Main Service (Business Logic)
	svc := service.New(rootArgs.service, logger, srcClient, mqttClient)

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
