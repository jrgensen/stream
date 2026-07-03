package jetstream

import (
	"github.com/nats-io/nats.go/jetstream"
)

type consumeContexts []jetstream.ConsumeContext

func (cc consumeContexts) Close() error {
	return nil
}

type subscription struct {
}

func (s *subscription) Close() error {
	return nil
}
