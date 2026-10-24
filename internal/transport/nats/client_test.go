package nats

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestConnectOptionsReconnectForever(t *testing.T) {
	opts := nats.GetDefaultOptions()
	for _, o := range connectOptions() {
		if err := o(&opts); err != nil {
			t.Fatalf("apply option: %v", err)
		}
	}
	if opts.MaxReconnect != -1 {
		t.Fatalf("MaxReconnect = %d, want -1 (unlimited)", opts.MaxReconnect)
	}
	if !opts.AllowReconnect {
		t.Fatal("AllowReconnect must stay enabled")
	}
	if opts.DisconnectedErrCB == nil || opts.ReconnectedCB == nil {
		t.Fatal("disconnect/reconnect handlers must be set for visibility")
	}
}
