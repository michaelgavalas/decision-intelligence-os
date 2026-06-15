package pubsub_test

import (
	"context"
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/pubsub"
)

func TestPublishSubscribeRoundTrip(t *testing.T) {
	pool := dbtest.NewPool(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := pubsub.NewSubscriber(pool)
	msgs, err := sub.Subscribe(ctx, "test_chan")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Give the LISTEN a brief moment to register before publishing.
	time.Sleep(100 * time.Millisecond)

	pub := pubsub.NewPublisher(pool)
	if err := pub.Publish(ctx, "test_chan", "hello"); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case got, ok := <-msgs:
		if !ok {
			t.Fatal("channel closed before delivering message")
		}
		if got != "hello" {
			t.Errorf("payload = %q, want %q", got, "hello")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for notification")
	}
}

func TestSubscribeChannelClosesOnCancel(t *testing.T) {
	pool := dbtest.NewPool(t)

	ctx, cancel := context.WithCancel(context.Background())

	sub := pubsub.NewSubscriber(pool)
	msgs, err := sub.Subscribe(ctx, "cancel_chan")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	cancel()

	select {
	case _, ok := <-msgs:
		if ok {
			t.Error("received a value, want closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("channel did not close after context cancellation")
	}
}
