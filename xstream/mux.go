package xstream

import (
	"context"

	"github.com/jrgensen/stream"
)

type mux struct {
	stream        stream.Stream
	opts          MuxOptions
	consumers     []stream.Consumer
	subscriptions []stream.Subscription
}

// The mux is a Stream message multiplexer. Like the standard http. ServeMux, mux. Router matches incoming requests against a list of registered routes and calls a handler for the route that matches the URL or other conditions.
func NewMux(stream stream.Stream, opts ...MuxOption) *mux {
	m := &mux{
		stream: stream,
	}
	for _, opt := range opts {
		opt(&m.opts)
	}
	return m
}

func (m *mux) AddConsumer(consumers ...stream.Consumer) {
	m.consumers = append(m.consumers, consumers...)
}

func (m *mux) Run(ctx context.Context) error {
	m.validate()
	if err := m.subscribe(); err != nil {
		return err
	}

	//if m.opts.blockUntilLive {
	//wait
	//}
	return nil
}

func (m *mux) validate() error {
	return nil
}

func (m *mux) subscribe() error {
	for _, consumer := range m.consumers {
		subscription, err := m.stream.Subscribe(consumer.Consumes(), consumer)
		if err != nil {
			return err
		}
		m.subscriptions = append(m.subscriptions, subscription)
	}
	return nil
}
