# subject

Package `subject` provides `StringSubject`, the canonical implementation of the
[`stream.Subject`](../README.md#subject) interface.

```go
import "github.com/jrgensen/stream/subject"
```

A subject addresses events and is composed of a **domain** and a **type**,
written in canonical form as `domain.type`.

## Constructors

```go
// From a canonical string. A single leading ":" is normalized to "."
// (so "orders:created" and "orders.created" are equivalent). Strings that
// contain whitespace are rejected and yield the zero subject.
s := subject.FromStr("orders.created")

// From explicit parts. An empty type produces a domain-only subject.
s := subject.FromParts("orders", "created")
```

Both return a `StringSubject` (a value type), which satisfies `stream.Subject`.

## Accessors

```go
s := subject.FromStr("orders.created")

s.Domain()   // "orders"       — everything before the first "."
s.Type()     // "created"      — everything after the first "."
s.Subject()  // "orders.created"
s.String()   // "orders.created"
s.Parts()    // []string{"orders", "created"}
```

## Matching

`Match` reports whether the subject matches a pattern. `.` is a literal
separator and `*` matches exactly one segment. Matching is case-insensitive.

```go
subject.FromStr("orders.created").Match("orders.*")   // true
subject.FromStr("orders.created").Match("orders.paid") // false
subject.FromStr("orders.created").Match("*.created")   // true
```

## Notes for implementors

- `StringSubject` is a small, comparable value (it deliberately opts out of `==`
  comparison via an internal `nocmp` field, so compare by `Subject()` instead).
- You are not required to use `StringSubject`. Any type implementing
  `stream.Subject` works; this package is provided so most callers don't have to
  write their own.
