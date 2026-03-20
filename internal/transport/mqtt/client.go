package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
}

func NewClient(broker string, clientID string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetResumeSubs(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	slog.Info("MQTT client initialized", "broker", broker, "client_id", clientID)
	return &Client{client: client}, nil
}

func (c *Client) PublishDiscovery(component string, objectID string, config interface{}) error {
	topic := fmt.Sprintf("homeassistant/%s/%s/config", component, objectID)
	payload, err := json.Marshal(config)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 0, true, payload)
	token.Wait()
	return token.Error()
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
