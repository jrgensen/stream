# xstream

Package `xstream` provides a message multiplexer that wires a set of
[`stream.Consumer`](../README.md#pipeline-metadata)s to a
[`stream.Stream`](../README.md#publish--subscribe).

```go
import "github.com/jrgensen/stream/xstream"
```

Like the standard library's `http.ServeMux`, the mux matches incoming messages
against registered consumers and dispatches each message to the consumer(s)
interested in its subject.

## Usage

```go
mux := xstream.NewMux(s) // s is a stream.Stream
mux.AddConsumer(ordersModel, auditModel)

if err := mux.Run(ctx); err != nil {
    return err
}
```

- `AddConsumer(...)` registers one or more `stream.Consumer`s. Each consumer
  advertises the subjects it wants via `Consumes()` and handles messages via
  `HandleMessage`.
- `Run(ctx)` subscribes every registered consumer to the underlying stream for
  the subjects it consumes, storing the resulting `stream.Subscription`s.

## Options

`NewMux` accepts functional options:

```go
mux := xstream.NewMux(s, xstream.MuxBlockUntilLive())
```

- `MuxBlockUntilLive()` sets the intent to block until the stream is live. See
  [`caughtup`](../caughtup) for the sentinel used to detect the "now live"
  point.

> Note: this package is an early building block; some routing/lifecycle
> behaviour (e.g. block-until-live) is still being fleshed out.
