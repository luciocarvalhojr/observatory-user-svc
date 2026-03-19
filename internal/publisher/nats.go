// Package publisher implements NATS event publishing for user-svc.
package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/domain"
)

const (
	SubjectUserCreated = "user.created"
	SubjectUserDeleted = "user.deleted"
)

// Publisher publishes domain events to NATS.
type Publisher struct {
	nc *nats.Conn
}

// New creates a new NATS Publisher.
func New(url string) (*Publisher, error) {
	nc, err := nats.Connect(url,
		nats.Name("observatory-user-svc"),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("publisher: connect to nats: %w", err)
	}
	return &Publisher{nc: nc}, nil
}

// Close drains and closes the NATS connection.
func (p *Publisher) Close() {
	_ = p.nc.Drain()
}

// UserCreated publishes a user.created event.
func (p *Publisher) UserCreated(_ context.Context, u *domain.User) error {
	return p.publish(SubjectUserCreated, domain.UserCreatedEvent{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	})
}

// UserDeleted publishes a user.deleted event.
func (p *Publisher) UserDeleted(_ context.Context, id string) error {
	return p.publish(SubjectUserDeleted, domain.UserDeletedEvent{ID: id})
}

func (p *Publisher) publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("publisher: marshal %s: %w", subject, err)
	}
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publisher: publish %s: %w", subject, err)
	}
	return nil
}
