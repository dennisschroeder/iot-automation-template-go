package nats

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestConnectOptionsReconnectForever(t *testing.T) {
	c := &Client{}
	opts := nats.GetDefaultOptions()
	for _, o := range c.connectOptions() {
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
	if opts.DisconnectedErrCB == nil || opts.ReconnectedCB == nil || opts.ClosedCB == nil {
		t.Fatal("disconnect/reconnect/closed handlers must be set")
	}
}

func TestOnClosedExitsUnlessClosedDeliberately(t *testing.T) {
	var code []int
	orig := exit
	exit = func(c int) { code = append(code, c) }
	t.Cleanup(func() { exit = orig })

	c := &Client{}
	c.onClosed(&nats.Conn{})
	if len(code) != 1 || code[0] != 1 {
		t.Fatalf("unexpected close must exit(1), got %v", code)
	}

	code = nil
	c.closing.Store(true)
	c.onClosed(&nats.Conn{})
	if len(code) != 0 {
		t.Fatalf("deliberate Close() must not exit, got %v", code)
	}
}
