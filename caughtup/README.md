# caughtup

Package `caughtup` provides a sentinel [`stream.Message`](../README.md#message--mutablemessage)
used to signal that a subscriber has replayed history and is now "live".

```go
import "github.com/jrgensen/stream/caughtup"
```

When a consumer reads an event store from the beginning, there is a moment where
it has caught up to the point the store was at when reading began. Emitting a
catch-up sentinel at that point lets downstream handlers switch behaviour — for
example, stop de-duplicating replayed events, or begin serving traffic.

See also [`stream.CatchupListener`](../README.md#pipeline-metadata).

## Usage

```go
// Produce a sentinel for a domain.
msg := caughtup.NewCaughtupMessage("orders") // subject: "orders.caughtup"

// Detect it in a handler.
func (m *model) HandleMessage(msg stream.Message) error {
    if caughtup.IsCaughtup(msg) {
        m.live = true
        return nil
    }
    // ... normal handling
    return nil
}
```

## API

- `CaughtupType` — the reserved subject type (`"caughtup"`).
- `NewCaughtupMessage(domain string) stream.Message` — build a sentinel whose
  subject is `domain.caughtup`. Its body and metadata are empty.
- `IsCaughtup(m stream.Message) bool` — true when a message's subject type is
  the catch-up sentinel.
