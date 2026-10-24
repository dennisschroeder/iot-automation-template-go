package mqtt

import (
	"errors"
	"sort"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type fakeToken struct{ err error }

func (t *fakeToken) Wait() bool                     { return true }
func (t *fakeToken) WaitTimeout(time.Duration) bool { return true }
func (t *fakeToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (t *fakeToken) Error() error { return t.err }

type fakeMessage struct {
	mqtt.Message
	topic   string
	payload []byte
}

func (m fakeMessage) Topic() string   { return m.topic }
func (m fakeMessage) Payload() []byte { return m.payload }

// fakeClient records SUBSCRIBEs; embedding the interface panics on any other
// call, which keeps the test honest about what resubscribe touches.
type fakeClient struct {
	mqtt.Client
	subscribed []string
	callbacks  map[string]mqtt.MessageHandler
	failTopics map[string]bool
}

func newFakeClient() *fakeClient {
	return &fakeClient{callbacks: map[string]mqtt.MessageHandler{}, failTopics: map[string]bool{}}
}

func (f *fakeClient) Subscribe(topic string, _ byte, cb mqtt.MessageHandler) mqtt.Token {
	f.subscribed = append(f.subscribed, topic)
	if f.failTopics[topic] {
		return &fakeToken{err: errors.New("subscribe refused")}
	}
	f.callbacks[topic] = cb
	return &fakeToken{}
}

func newTestClient(f *fakeClient) *Client {
	return &Client{client: f, subs: make(map[string]MessageHandler)}
}

func TestResubscribeOnReconnect(t *testing.T) {
	f := newFakeClient()
	c := newTestClient(f)

	c.resubscribe(f) // initial connect
	if len(f.subscribed) != 0 {
		t.Fatalf("initial connect must not subscribe, got %v", f.subscribed)
	}

	var gotTopic, gotPayload string
	if err := c.Subscribe("svc/switch/+/set", func(topic string, payload []byte) {
		gotTopic, gotPayload = topic, string(payload)
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Simulate the broker having expired the session: the client reconnects
	// and must subscribe again, on every reconnect.
	for i := 0; i < 2; i++ {
		f.subscribed = nil
		delete(f.callbacks, "svc/switch/+/set")
		c.resubscribe(f)
		if len(f.subscribed) != 1 || f.subscribed[0] != "svc/switch/+/set" {
			t.Fatalf("reconnect %d: expected resubscribe to svc/switch/+/set, got %v", i+1, f.subscribed)
		}
	}

	cb, ok := f.callbacks["svc/switch/+/set"]
	if !ok {
		t.Fatal("no callback registered on resubscribe")
	}
	cb(f, fakeMessage{topic: "svc/switch/lock_x/set", payload: []byte("OFF")})
	if gotTopic != "svc/switch/lock_x/set" || gotPayload != "OFF" {
		t.Fatalf("handler got topic=%q payload=%q", gotTopic, gotPayload)
	}
}

func TestResubscribeContinuesAfterFailure(t *testing.T) {
	f := newFakeClient()
	c := newTestClient(f)
	c.resubscribe(f)

	noop := func(string, []byte) {}
	for _, topic := range []string{"a/set", "b/set"} {
		if err := c.Subscribe(topic, noop); err != nil {
			t.Fatalf("Subscribe %s: %v", topic, err)
		}
	}

	f.subscribed = nil
	f.failTopics["a/set"] = true
	c.resubscribe(f)

	sort.Strings(f.subscribed)
	if len(f.subscribed) != 2 || f.subscribed[0] != "a/set" || f.subscribed[1] != "b/set" {
		t.Fatalf("expected both topics attempted, got %v", f.subscribed)
	}
	if _, ok := f.callbacks["b/set"]; !ok {
		t.Fatal("b/set not resubscribed after a/set failed")
	}
}
