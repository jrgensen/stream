# metatagger

Package `metatagger` provides a [`stream.Publisher`](../README.md#publish--subscribe)
decorator that merges a set of **default metadata** fields into every message
before delegating to a wrapped publisher.

```go
import "github.com/jrgensen/stream/metatagger"
```

It removes the repetitive need to set the same metadata (for example a
`producer` tag) on every message: configure the defaults once, then publish
normally.

## Usage

```go
tagger, err := metatagger.New(next, Metadata{Producer: "tilmelding-api"})
if err != nil {
    return err
}

msg := tagger.MessageFunc()(subject.FromStr("orders.created"))
msg.SetBody(&order)
tagger.Publish(msg) // meta.producer is filled in automatically
```

`next` is any `stream.Publisher` (e.g. a [`jetstream`](../jetstream) stream). The
tagger implements `stream.Publisher` itself, so it drops in wherever a publisher
is expected.

## Semantics

- **Per-message metadata always wins.** A default is applied only for a field
  the message does not already provide, so callers can override individual
  fields while inheriting the rest.
- **Defaults are any JSON object.** Pass a struct (such as your `Metadata`
  type), a `map[string]…`, or anything that marshals to a JSON object. A `nil`
  defaults value makes the tagger a transparent pass-through.
- **Read-only messages pass through untouched.** If a message does not
  implement `stream.MutableMessage`, it is forwarded unchanged rather than
  dropped.
- **`MessageFunc()` delegates** to the wrapped publisher, so constructed
  messages remain compatible with the underlying transport.

`New` returns an error if the supplied defaults cannot be marshalled to a JSON
object.
