package mqtt

import (
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
}

type MessageHandler func(topic string, payload []byte)

func NewClient(broker, clientID string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	slog.Info("MQTT client connected", "broker", broker)
	return &Client{client: client}, nil
}

func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	token.Wait()
	return token.Error()
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
