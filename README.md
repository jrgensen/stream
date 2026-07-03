# stream

`stream` defines a small publisher–subscriber messaging interface and the
related types used to compose event-driven data-pipelines in Go.

The root package is **interfaces only** — the contract. Concrete building
blocks (a NATS JetStream transport, an in-memory test double, a metadata
decorator, a subject implementation, …) live in sub-packages and can be mixed
and matched, or replaced entirely by your own implementations.

```go
import "github.com/jrgensen/stream"
```

- Module path: `github.com/jrgensen/stream`
- Go: 1.25+

## Why interfaces?

Reading remote messages from a broker (NATS) or persisting state in a database
is fundamentally different from the event-passing you do inside a service. The
`stream` interfaces let you keep those concerns in separate layers, each of
which only depends on the contract — never on a concrete transport or codec.

```
  NATS               |
 ------------------  |  Knows about NATS *and* stream. Data abstraction layer.
  nats client        |
 ------------------  —
  stream router      |
 ------------------  |  Knows only about stream interfaces.
  stream handlers    |
 ------------------  —
  db handlers        |
 ------------------  |  Knows about MongoDB *and* stream. Data abstraction layer.
  MongoDB            |
```

Rules of thumb this design encourages:

- **Decode and validate at the source boundary.** Once a message is inside the
  pipeline it is trusted; internal handlers should not re-parse transport
  encodings such as JSON.
- **Business logic works on `stream.Message`, not on transport types.** The
  router and handlers must not know about NATS or the wire codec.
- **Encoding for storage is the storage handler's job.** The core packages know
  nothing about the databases that implement their interfaces.

## The contract

Everything below lives in the root `stream` package.

### Message / MutableMessage

`Message` is a **read-only** view of a single event on the stream.

```go
type Message interface {
    Subject() Subject          // where the event was published
    Time() time.Time           // when it occurred / was submitted
    Sequence() uint64          // monotonic position within the stream
    Body(dst interface{}) error // decode the value into dst (a struct pointer)
    Meta(dst interface{}) error // decode optional metadata into dst
    RawBody() interface{}       // the undecoded value
    RawMeta() interface{}       // the undecoded metadata
}
```

Implementor notes:

- `Body`/`Meta` copy the payload into `dst`. `dst` is borrowed only for the
  duration of the call. How the copy happens is up to you — decode a stored
  byte slice, deep-copy an in-memory struct, etc.
- `RawBody`/`RawMeta` expose the underlying representation for consumers (such
  as decorators) that want to move bytes around without decoding. Document the
  concrete type you return; existing helpers accept `json.RawMessage`, `[]byte`,
  `string`, or an arbitrary value.

`MutableMessage` adds the setters used while a message is being built:

```go
type MutableMessage interface {
    Message
    SetSubject(Subject)
    SetBody(interface{}) error
    SetMeta(interface{}) error
    SetTime(time.Time) error
}
```

`MessageFunc` is the factory that produces new mutable messages for a subject.
Every `Publisher` exposes one so callers construct messages that are compatible
with the underlying transport:

```go
type MessageFunc func(Subject) MutableMessage
```

### Handling messages

```go
type MessageHandler interface {
    HandleMessage(Message) error
}
```

`MessageHandlerFunc` is the idiomatic function adapter (à la `http.HandlerFunc`):

```go
var h stream.MessageHandler = stream.MessageHandlerFunc(func(m stream.Message) error {
    // ...
    return nil
})
```

### Publish / Subscribe

```go
type Publisher interface {
    Publish(msg Message) error   // may be sync or async; must be thread-safe
    MessageFunc() MessageFunc    // factory for transport-compatible messages
}

type Subscriber interface {
    // Subscribe registers h for the given subjects and returns a Subscription.
    // Call Close() on the subscription to unsubscribe.
    Subscribe(subjects []Subject, h MessageHandler) (Subscription, error)
}

type Subscription interface {
    Close() error
}

// Stream is the full duplex contract: a Publisher + Subscriber you can Close.
type Stream interface {
    Publisher
    Subscriber
    Close() error
}
```

`PublisherFunc` adapts a bare `func(Message) error` to `Publisher` (its
`MessageFunc` panics — use it only where a factory is not required).

### Pipeline metadata

These optional interfaces help wire a graph of publishers and subscribers, and
let handlers react to stream lifecycle events.

```go
// Producer advertises the subjects a component may write to.
type Producer interface {
    Produces() []Subject
}

// Consumer is a MessageHandler that advertises the subjects it wants.
type Consumer interface {
    MessageHandler
    Consumes() []Subject
}

// CatchupListener is called once a handler has replayed history up to the
// point the store was at when reading began — useful for de-duplication and
// other "we are now live" optimizations.
type CatchupListener interface {
    CaughtUp()
}
```

### Subject

A `Subject` addresses events and is composed of a `domain` and a `type`.

```go
type Subject interface {
    Domain() string       // the domain part
    Type() string         // the type part
    Subject() string      // canonical "domain.type" string
    Parts() []string      // split on "."
    Match(string) bool    // pattern match (supports "*" wildcards)
}
```

The canonical implementation lives in [`subject`](./subject).

## Composing programs

Internal services are built as data-pipelines. An input (usually from NATS) is
consumed and processed by one or more subscribers. Each subscriber transforms
data and publishes its output to a shared in-process `Stream`; other handlers
subscribe to that stream and repeat the process.

```
  [handler] --||-- [handler] --||-- [handler]
```

The `||` between handlers is a routing component (see [`xstream`](./xstream))
responsible for delivering messages from publishers to subscribers. When the
originating NATS event is published, only the first subscriber is known; that
subscriber has no knowledge of downstream consumers of the messages it produces.

## Packages

| Package | Kind | Description |
| --- | --- | --- |
| `stream` (root) | interfaces | The messaging contract and function adapters. |
| [`subject`](./subject) | implementation | Canonical `Subject` (`StringSubject`, `FromStr`, `FromParts`). |
| [`jetstream`](./jetstream) | implementation | `Stream` backed by NATS JetStream. |
| [`metatagger`](./metatagger) | implementation | `Publisher` decorator that applies default metadata. |
| [`xstream`](./xstream) | implementation | Message multiplexer/router wiring `Consumer`s to a `Stream`. |
| [`caughtup`](./caughtup) | implementation | Sentinel `Message` signalling catch-up state. |
| [`streamtest`](./streamtest) | implementation | In-memory `Publisher`/`Message` doubles for tests. |

## Implementing your own transport

To add a new backend, implement `stream.Stream`:

1. **`MessageFunc()`** — return a factory that builds `MutableMessage` values
   your `Publish` can serialize. Keeping construction and publishing in the same
   implementation lets you attach whatever bookkeeping (IDs, versions) you need.
2. **`Publish(Message)`** — serialize and send. Must be safe for concurrent use.
3. **`Subscribe([]Subject, MessageHandler)`** — deliver matching messages to the
   handler and return a `Subscription` whose `Close` tears the subscription down.
4. **`Close()`** — stop accepting new messages and unsubscribe everything.

Use [`jetstream`](./jetstream) as a reference and [`streamtest`](./streamtest)
to exercise handlers without any external infrastructure.

## License

Released under the [MIT License](./LICENSE).
