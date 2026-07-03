# jetstream

Package `jetstream` implements [`stream.Stream`](../README.md#publish--subscribe)
on top of [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream), giving
you a durable, ordered event transport behind the `stream` interfaces.

```go
import "github.com/jrgensen/stream/jetstream"
```

## Usage

```go
s, err := jetstream.New(nats.DefaultURL) // "" also defaults to the local server
if err != nil {
    return err
}
defer s.Close()

// Build a transport-compatible message and publish it.
msg := s.MessageFunc()(subject.FromStr("orders.created"))
msg.SetBody(&Order{ID: 42})
if err := s.Publish(msg); err != nil {
    return err
}

// Subscribe. Handlers receive decoded stream.Message values.
sub, err := s.Subscribe(
    []stream.Subject{subject.FromStr("orders.created")},
    stream.MessageHandlerFunc(func(m stream.Message) error {
        var o Order
        return m.Body(&o)
    }),
)
defer sub.Close()
```

## Behaviour

- **Subjects → NATS subjects.** `Publish` maps a `stream.Subject` onto a NATS
  subject of the form `DOMAIN.type` (the domain is upper-cased). `Subscribe`
  groups requested subjects by domain and creates an ordered consumer per
  domain.
- **Envelope.** Messages are wrapped in a JSON envelope (`envelope.go`) carrying
  `eventId`, `correlationId`, `causationId`, `version`, `time`, `body` and
  `meta`. `New`-created messages get a fresh event ID that is also used as the
  correlation and causation ID until you override it.
- **Identity.** Published messages must implement the package's `Identifiable`
  interface (`EventID`/`CorrelationID`/`CausationID`); the message type returned
  by `MessageFunc()`/`NewMessage()` already does. Use
  `SetCausationCorrelationFromMessage` to chain a new message to the one that
  caused it.
- **`LastMessage(subject)`** fetches the most recent message for a subject via an
  ordered consumer with a last-delivery policy.

## Requirements

A reachable NATS server with JetStream enabled. The stream named `NATHEJK` is
referenced by the consumer setup; use `Create(name)` to create a JetStream
stream (`name.>`) if you need to provision one.

For tests and local development that should not depend on a broker, use the
in-memory [`streamtest`](../streamtest) doubles instead.
