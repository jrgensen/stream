package streamtest

import (
	"time"

	json "github.com/json-iterator/go"

	"github.com/jrgensen/stream"
	"github.com/jrgensen/stream/subject"
)

type SingleDomainPublisher chan stream.Message

func (s *SingleDomainPublisher) Publish(msg stream.Message) error {
	*s <- msg
	return nil
}

func (s *SingleDomainPublisher) MessageFunc() stream.MessageFunc {
	return MessageFunc
}

func (s *SingleDomainPublisher) Pop() (msg stream.Message, exists bool) {
	select {
	case msg = <-*s:
		exists = true
	default:
		exists = false
	}
	return
}

var _ stream.Publisher = (*SingleDomainPublisher)(nil)

type Message struct {
	time time.Time
	seq  uint64
	body []byte
	meta []byte
	subj stream.Subject
}

func NewMessage(subject stream.Subject) *Message {
	return &Message{
		subj: subject,
	}
}

func MessageFunc(subject stream.Subject) stream.MutableMessage {
	return NewMessage(subject)
}

type MessageData struct {
	Time time.Time
	Body interface{}
	Meta interface{}
}

func NewMessageP(subject stream.Subject, opts MessageData) *Message {
	m := NewMessage(subject)
	if err := m.SetBody(opts.Body); err != nil {
		panic(err)
	}
	if err := m.SetMeta(opts.Meta); err != nil {
		panic(err)
	}
	if err := m.SetTime(opts.Time); err != nil {
		panic(err)
	}
	return m
}

func (m *Message) SetSubject(subj stream.Subject) {
	m.subj = subj
}

func (m *Message) SetBody(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.body = b
	return nil
}

func (m *Message) SetMeta(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.meta = b
	return nil
}

func (m *Message) SetTime(t time.Time) error {
	m.time = t
	return nil
}

func (m *Message) Subject() stream.Subject    { return m.subj }
func (m *Message) Time() time.Time            { return m.time }
func (m *Message) Sequence() uint64           { return m.seq }
func (m *Message) Body(dst interface{}) error { return json.Unmarshal(m.body, dst) }
func (m *Message) Meta(dst interface{}) error { return json.Unmarshal(m.meta, dst) }
func (m *Message) RawBody() interface{}       { return m.body }
func (m *Message) RawMeta() interface{}       { return m.meta }

func StubBody(domain, typ string, body interface{}) []stream.Message {
	return []stream.Message{NewMessageP(subject.FromParts(domain, typ), MessageData{
		Body: body,
	})}
}

func SeedModel(model stream.MessageHandler, msgs ...[]stream.Message) {
	for _, msg := range msgs {
		for _, m := range msg {
			model.HandleMessage(m)
		}
	}
}

var _ stream.MessageFunc = MessageFunc
var _ stream.Message = (*Message)(nil)
var _ stream.MutableMessage = (*Message)(nil)
