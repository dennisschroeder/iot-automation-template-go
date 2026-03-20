package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Config represents MQTT connection settings
type Config struct {
	Host     string
	Port     int
	ClientID string
}

// Client wraps the Paho MQTT client
type Client struct {
	config Config
	logger *slog.Logger
	client paho.Client
}

// NewClient initializes a new MQTT client
func NewClient(cfg Config, logger *slog.Logger) (*Client, error) {
	opts := paho.NewClientOptions()
	brokerUrl := fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port)
	opts.AddBroker(brokerUrl)
	opts.SetClientID(cfg.ClientID)

	opts.SetOnConnectHandler(func(c paho.Client) {
		logger.Info("Connected to MQTT broker", "broker", brokerUrl)
	})
	opts.SetConnectionLostHandler(func(c paho.Client, err error) {
		logger.Warn("MQTT connection lost", "error", err)
	})

	return &Client{
		config: cfg,
		logger: logger,
		client: paho.NewClient(opts),
	}, nil
}

// Connect attempts to connect to the broker, retrying indefinitely
func (m *Client) Connect(ctx context.Context) error {
	m.logger.Info("Connecting to MQTT broker...")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if token := m.client.Connect(); token.Wait() && token.Error() == nil {
				return nil
			} else {
				m.logger.Error("Failed to connect to MQTT broker, retrying...", "error", token.Error())
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (m *Client) Disconnect() {
	m.client.Disconnect(250)
	m.logger.Info("Disconnected from MQTT broker")
}

// PublishDiscovery simplifies Home Assistant MQTT Auto-Discovery
func (m *Client) PublishDiscovery(sensorID, name, stateClass, deviceClass, unit string) {
	topic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", m.config.ClientID, sensorID)
	payload := map[string]interface{}{
		"name":        name,
		"state_topic": fmt.Sprintf("%s/sensor/%s/state", m.config.ClientID, sensorID),
		"unique_id":   fmt.Sprintf("%s_%s", m.config.ClientID, sensorID),
		"device": map[string]interface{}{
			"identifiers":  []string{m.config.ClientID},
			"name":         m.config.ClientID,
			"manufacturer": "Homelab Custom",
		},
	}

	if stateClass != "" {
		payload["state_class"] = stateClass
	}
	if deviceClass != "" {
		payload["device_class"] = deviceClass
	}
	if unit != "" {
		payload["unit_of_measurement"] = unit
	}

	jsonData, _ := json.Marshal(payload)
	m.client.Publish(topic, 0, true, string(jsonData))
}

// PublishState publishes a sensor reading to the state topic
func (m *Client) PublishState(sensorID string, value interface{}) {
	topic := fmt.Sprintf("%s/sensor/%s/state", m.config.ClientID, sensorID)
	m.client.Publish(topic, 0, false, fmt.Sprintf("%v", value))
}
