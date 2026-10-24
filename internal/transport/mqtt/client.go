package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client

	mu            sync.Mutex
	subs          map[string]MessageHandler
	connectedOnce bool
}

func NewClient(broker string, clientID string) (*Client, error) {
	c := &Client{subs: make(map[string]MessageHandler)}
	c.client = mqtt.NewClient(c.options(broker, clientID))
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	slog.Info("MQTT client initialized", "broker", broker, "client_id", clientID)
	return c, nil
}

func (c *Client) options(broker, clientID string) *mqtt.ClientOptions {
	return mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetResumeSubs(true).
		SetOnConnectHandler(c.resubscribe).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			slog.Warn("MQTT connection lost", "error", err)
		})
}

type MessageHandler func(topic string, payload []byte)

// subscribeTimeout bounds the wait for a SUBACK. If the connection drops
// mid-subscribe, ResumeSubs delivers the SUBSCRIBE on the next connection.
const subscribeTimeout = 10 * time.Second

// resubscribe runs on every (re)connect. ResumeSubs only resends SUBSCRIBEs
// that were never acknowledged, so once the broker has dropped the persistent
// session (Mosquitto: persistent_client_expiration) the client would come back
// connected but without any subscriptions.
func (c *Client) resubscribe(client mqtt.Client) {
	c.mu.Lock()
	// The initial connect has nothing to restore; Subscribe does that.
	if !c.connectedOnce {
		c.connectedOnce = true
		c.mu.Unlock()
		return
	}
	subs := make(map[string]MessageHandler, len(c.subs))
	for topic, handler := range c.subs {
		subs[topic] = handler
	}
	c.mu.Unlock()

	if len(subs) == 0 {
		return
	}
	slog.Info("MQTT reconnected, resubscribing", "topics", len(subs))
	for topic, handler := range subs {
		if err := subscribe(client, topic, handler); err != nil {
			slog.Error("MQTT resubscribe failed", "topic", topic, "error", err)
		}
	}
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

func (c *Client) Publish(topic string, payload interface{}) error {
	token := c.client.Publish(topic, 0, false, payload)
	token.Wait()
	return token.Error()
}

func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	c.mu.Lock()
	c.subs[topic] = handler
	c.mu.Unlock()
	if err := subscribe(c.client, topic, handler); err != nil {
		slog.Error("MQTT subscribe failed", "topic", topic, "error", err)
		return err
	}
	return nil
}

func subscribe(client mqtt.Client, topic string, handler MessageHandler) error {
	token := client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	if !token.WaitTimeout(subscribeTimeout) {
		return fmt.Errorf("no SUBACK within %s", subscribeTimeout)
	}
	return token.Error()
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
