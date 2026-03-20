package watcher

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Watcher interface {
	Subject() string
	Queue() string
	Handle(msg *nats.Msg)
}

type Manager struct {
	js nats.JetStreamContext
}

func NewManager(js nats.JetStreamContext) *Manager {
	return &Manager{js: js}
}

func (m *Manager) Register(w Watcher) error {
	slog.Info("Registering watcher", "subject", w.Subject(), "queue", w.Queue())

	var sub *nats.Subscription
	var err error

	if w.Queue() != "" {
		sub, err = m.js.QueueSubscribe(w.Subject(), w.Queue(), w.Handle)
	} else {
		sub, err = m.js.Subscribe(w.Subject(), w.Handle)
	}

	if err != nil {
		return err
	}

	slog.Debug("Watcher registered", "subject", sub.Subject)
	return nil
}
