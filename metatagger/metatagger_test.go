package metatagger_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jrgensen/stream"
	"github.com/jrgensen/stream/metatagger"
	"github.com/jrgensen/stream/streamtest"
	"github.com/jrgensen/stream/subject"
)

type meta struct {
	Producer string `json:"producer,omitempty"`
	Actor    string `json:"actor,omitempty"`
}

// capture is a Publisher that records the last message it received.
type capture struct {
	last stream.Message
	err  error
}

func (c *capture) Publish(msg stream.Message) error {
	c.last = msg
	return c.err
}

func (c *capture) MessageFunc() stream.MessageFunc {
	return streamtest.MessageFunc
}

func decodeMeta(t *testing.T, msg stream.Message) meta {
	t.Helper()
	var m meta
	if err := msg.Meta(&m); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	return m
}

func newPub(t *testing.T, next stream.Publisher, defaults interface{}) *metatagger.Publisher {
	t.Helper()
	pub, err := metatagger.New(next, defaults)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return pub
}

func TestAppliesDefaultsWhenMetaAbsent(t *testing.T) {
	sink := &capture{}
	pub := newPub(t, sink, meta{Producer: "tilmelding-api"})

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	if err := pub.Publish(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := decodeMeta(t, sink.last)
	if got.Producer != "tilmelding-api" {
		t.Fatalf("producer = %q, want %q", got.Producer, "tilmelding-api")
	}
}

func TestPerMessageMetaWins(t *testing.T) {
	sink := &capture{}
	pub := newPub(t, sink, meta{Producer: "tilmelding-api"})

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	if err := msg.SetMeta(meta{Producer: "migrate-orders"}); err != nil {
		t.Fatalf("set meta: %v", err)
	}
	if err := pub.Publish(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := decodeMeta(t, sink.last)
	if got.Producer != "migrate-orders" {
		t.Fatalf("producer = %q, want %q (per-message should win)", got.Producer, "migrate-orders")
	}
}

func TestFieldsAreMergedIndependently(t *testing.T) {
	sink := &capture{}
	pub := newPub(t, sink, meta{Producer: "tilmelding-api"})

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	// Only set Actor; Producer should be inherited from the defaults.
	if err := msg.SetMeta(meta{Actor: "user-42"}); err != nil {
		t.Fatalf("set meta: %v", err)
	}
	if err := pub.Publish(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := decodeMeta(t, sink.last)
	if got.Producer != "tilmelding-api" {
		t.Fatalf("producer = %q, want inherited default %q", got.Producer, "tilmelding-api")
	}
	if got.Actor != "user-42" {
		t.Fatalf("actor = %q, want %q", got.Actor, "user-42")
	}
}

func TestNilDefaultsIsTransparent(t *testing.T) {
	sink := &capture{}
	pub := newPub(t, sink, nil)

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	if err := msg.SetMeta(meta{Producer: "explicit"}); err != nil {
		t.Fatalf("set meta: %v", err)
	}
	if err := pub.Publish(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := decodeMeta(t, sink.last)
	if got.Producer != "explicit" {
		t.Fatalf("producer = %q, want %q", got.Producer, "explicit")
	}
}

func TestPublishErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	sink := &capture{err: sentinel}
	pub := newPub(t, sink, meta{Producer: "tilmelding-api"})

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	if err := pub.Publish(msg); !errors.Is(err, sentinel) {
		t.Fatalf("publish err = %v, want %v", err, sentinel)
	}
}

func TestInvalidDefaultsReturnError(t *testing.T) {
	sink := &capture{}
	// A slice does not marshal to a JSON object.
	if _, err := metatagger.New(sink, []string{"nope"}); err == nil {
		t.Fatal("expected error for non-object defaults, got nil")
	}
}

func TestDefaultsAcceptMap(t *testing.T) {
	sink := &capture{}
	pub := newPub(t, sink, map[string]string{"producer": "from-map"})

	msg := pub.MessageFunc()(subject.FromStr("NATHEJK.crewmember.created"))
	if err := pub.Publish(msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	raw, ok := sink.last.RawMeta().([]byte)
	if !ok {
		t.Fatalf("raw meta type = %T, want []byte", sink.last.RawMeta())
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["producer"] != "from-map" {
		t.Fatalf("producer = %q, want %q", m["producer"], "from-map")
	}
}
