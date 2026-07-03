package caughtup

import (
	"time"

	"github.com/jrgensen/stream"
	"github.com/jrgensen/stream/subject"
)

// CaughtupType is a sentinel used to communicate 'caughtup' for streams.
const CaughtupType = "caughtup"

type caughtup struct {
	t    time.Time
	subj stream.Subject
}

func (m caughtup) Subject() stream.Subject { return m.subj }
func (m caughtup) Time() time.Time         { return m.t }
func (m caughtup) Sequence() uint64        { return 0 }
func (m caughtup) Body(interface{}) error  { return nil }
func (m caughtup) Meta(interface{}) error  { return nil }
func (m caughtup) RawBody() interface{}    { return nil }
func (m caughtup) RawMeta() interface{}    { return nil }

// NewCaughtupMessage creates a new 'caughtup' message. This is used as an
// internal senitel message to communicate 'caughtup' state.
func NewCaughtupMessage(domain string) stream.Message {
	return caughtup{subj: subject.FromParts(domain, CaughtupType), t: time.Now().UTC()}
}

// IsCaughtup is true if the message is an the internal 'caughtup' senitel.
func IsCaughtup(m stream.Message) bool {
	return m.Subject().Type() == CaughtupType
}

var _ stream.Message = (*caughtup)(nil)
