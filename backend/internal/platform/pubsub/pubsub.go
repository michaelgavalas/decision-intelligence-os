// Package pubsub provides lightweight publish/subscribe messaging built on
// PostgreSQL LISTEN/NOTIFY. It lets the backend fan out real-time updates (for
// GraphQL subscriptions) without introducing a separate message broker.
package pubsub

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	platformerrors "github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// subscribeBuffer is the capacity of the delivery channel. Buffering smooths
// short consumer stalls without unbounded memory growth.
const subscribeBuffer = 16

// Publisher sends notifications on a channel using pg_notify.
type Publisher struct {
	pool *pgxpool.Pool
}

// NewPublisher returns a Publisher backed by the given pool.
func NewPublisher(pool *pgxpool.Pool) *Publisher {
	return &Publisher{pool: pool}
}

// Publish sends payload on the given channel. The channel name is passed as a
// bind parameter to pg_notify, so it does not require manual quoting here.
func (p *Publisher) Publish(ctx context.Context, channel, payload string) error {
	if _, err := p.pool.Exec(ctx, "SELECT pg_notify($1, $2)", channel, payload); err != nil {
		return platformerrors.Wrap(err, platformerrors.KindInternal, "PUBSUB_PUBLISH_FAILED", "failed to publish notification")
	}
	return nil
}

// Subscriber listens for notifications on a channel.
type Subscriber struct {
	pool *pgxpool.Pool
}

// NewSubscriber returns a Subscriber backed by the given pool.
func NewSubscriber(pool *pgxpool.Pool) *Subscriber {
	return &Subscriber{pool: pool}
}

// Subscribe acquires a dedicated connection, issues LISTEN on the sanitized
// channel, and returns a receive-only channel of payloads. A background
// goroutine forwards notifications until ctx is cancelled, at which point it
// releases the connection and closes the returned channel.
func (s *Subscriber) Subscribe(ctx context.Context, channel string) (<-chan string, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, platformerrors.Wrap(err, platformerrors.KindInternal, "PUBSUB_ACQUIRE_FAILED", "failed to acquire connection")
	}

	// LISTEN cannot be parameterized, so the identifier must be quoted safely.
	listen := "LISTEN " + pgx.Identifier{channel}.Sanitize()
	if _, err := conn.Exec(ctx, listen); err != nil {
		conn.Release()
		return nil, platformerrors.Wrap(err, platformerrors.KindInternal, "PUBSUB_LISTEN_FAILED", "failed to listen on channel")
	}

	out := make(chan string, subscribeBuffer)
	go s.forward(ctx, conn, out)

	return out, nil
}

// forward delivers notifications from the connection to out until ctx is
// cancelled or the connection fails, then cleans up.
func (s *Subscriber) forward(ctx context.Context, conn *pgxpool.Conn, out chan<- string) {
	defer close(out)
	defer conn.Release()

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			// A cancelled context is the normal shutdown path; any other error
			// means the connection is no longer usable, so stop forwarding.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			return
		}

		select {
		case out <- notification.Payload:
		case <-ctx.Done():
			return
		}
	}
}
