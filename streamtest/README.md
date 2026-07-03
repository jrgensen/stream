# streamtest

Package `streamtest` provides in-memory doubles for the `stream` interfaces so
you can exercise handlers, models and publishers without any external
infrastructure (no NATS, no database).

```go
import "github.com/jrgensen/stream/streamtest"
```

## What's inside

### `Message`

A concrete, JSON-backed [`stream.MutableMessage`](../README.md#message--mutablemessage).

```go
msg := streamtest.NewMessage(subject.FromStr("orders.created"))
msg.SetBody(&order)

// Or in one call:
msg := streamtest.NewMessageP(subject.FromStr("orders.created"), streamtest.MessageData{
    Time: time.Now(),
    Body: order,
    Meta: meta,
})
```

`MessageFunc` is a ready-made `stream.MessageFunc` that builds these messages —
handy when a component asks a publisher for its factory.

### `SingleDomainPublisher`

A channel-backed [`stream.Publisher`](../README.md#publish--subscribe). Published
messages are buffered on the channel and can be drained in assertions.

```go
pub := make(streamtest.SingleDomainPublisher, 8)
pub.Publish(msg)

got, ok := pub.Pop() // non-blocking; ok == false when empty
```

### Seeding models

`StubBody` and `SeedModel` make it easy to feed a `stream.MessageHandler`
(such as a read model) a sequence of events:

```go
streamtest.SeedModel(model,
    streamtest.StubBody("orders", "created", CreatedPayload{ID: 1}),
    streamtest.StubBody("orders", "paid", PaidPayload{ID: 1}),
)
```

- `StubBody(domain, typ, body)` returns a one-element `[]stream.Message` with the
  given subject and body.
- `SeedModel(handler, msgs...)` calls `HandleMessage` for every message, in
  order.

## Notes

- Bodies and metadata are marshalled with `json-iterator`, matching the JSON
  semantics of the production transports, so type round-tripping behaves the
  same as it does over the wire.
- This package is intended for use in `_test.go` files.
