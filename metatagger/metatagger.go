// Package metatagger provides a stream.Publisher decorator that
// applies default metadata values to every message before delegating
// publication to a wrapped Publisher.
//
// It removes the repetitive need to set the same metadata (for example a
// "producer" tag) on every message. Instead you configure the defaults once
// when constructing the tagger:
//
//	tagger, err := metatagger.New(jetstream, messages.Metadata{Producer: "tilmelding-api"})
//	if err != nil {
//		return err
//	}
//	msg := tagger.MessageFunc()(subject)
//	msg.SetBody(&body)
//	tagger.Publish(msg) // meta is filled in with the defaults automatically
//
// Per-message metadata always wins: a default is only applied for a field
// that the message does not already provide. This lets callers override
// individual fields while inheriting the rest.
package metatagger

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/jrgensen/stream"
)

// Publisher wraps another stream.Publisher and merges a set of
// default metadata fields into every published message.
type Publisher struct {
	next     stream.Publisher
	defaults map[string]json.RawMessage
}

var _ stream.Publisher = (*Publisher)(nil)

// New returns a Publisher that decorates next, applying the given default
// metadata to every message it publishes.
//
// defaults may be any value that marshals to a JSON object (typically a
// struct such as messages.Metadata, or a map). A nil defaults leaves the
// tagger as a transparent pass-through. New returns an error if defaults
// cannot be marshalled to a JSON object.
func New(next stream.Publisher, defaults interface{}) (*Publisher, error) {
	obj, err := toObject(defaults)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		next:     next,
		defaults: obj,
	}, nil
}

// Publish merges the configured default metadata into msg and forwards it to
// the wrapped Publisher. Fields already present on the message's metadata are
// left untouched.
func (p *Publisher) Publish(msg stream.Message) error {
	if len(p.defaults) == 0 {
		return p.next.Publish(msg)
	}

	mutable, ok := msg.(stream.MutableMessage)
	if !ok {
		// We cannot tag a read-only message; forward it unchanged rather
		// than dropping it.
		return p.next.Publish(msg)
	}

	merged, err := p.merge(msg.RawMeta())
	if err != nil {
		return err
	}
	if err := mutable.SetMeta(merged); err != nil {
		return err
	}
	return p.next.Publish(mutable)
}

// MessageFunc delegates to the wrapped Publisher so callers construct messages
// compatible with the underlying stream.
func (p *Publisher) MessageFunc() stream.MessageFunc {
	return p.next.MessageFunc()
}

// merge builds the metadata object that should be published for a message with
// the given raw metadata. It starts from the defaults and overlays any fields
// the message already set.
func (p *Publisher) merge(rawMeta interface{}) (map[string]json.RawMessage, error) {
	out := make(map[string]json.RawMessage, len(p.defaults))
	for k, v := range p.defaults {
		out[k] = v
	}

	raw := rawBytes(rawMeta)
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return out, nil
	}

	msgMeta := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &msgMeta); err != nil {
		return nil, err
	}
	for k, v := range msgMeta {
		out[k] = v
	}
	return out, nil
}

// toObject marshals v and decodes it back into a field map. Values that are
// nil or do not represent a JSON object yield a nil map (transparent
// pass-through). It returns an error if v cannot be marshalled to a JSON
// object.
func toObject(v interface{}) (map[string]json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	raw := rawBytes(v)
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	obj := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("metatagger: defaults must marshal to a JSON object: %w", err)
	}
	if len(obj) == 0 {
		return nil, nil
	}
	return obj, nil
}

// rawBytes normalises the various concrete types RawMeta/RawBody may return
// into JSON bytes.
func rawBytes(v interface{}) []byte {
	switch t := v.(type) {
	case nil:
		return nil
	case json.RawMessage:
		return []byte(t)
	case []byte:
		return t
	case string:
		return []byte(t)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return b
	}
}
