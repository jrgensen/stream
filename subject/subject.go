// Package subject provides the canonical StringSubject implementation of the
// stream.Subject interface.
package subject

import (
	"regexp"
	"strings"
)

// StringSubject is the canonical implementation of the stream.Subject
// interface.
type StringSubject struct {
	_ [0]func() // nocmp
	i uint16
	j uint16
	s string
}

func FromStr(s string) StringSubject {
	if strings.ContainsAny(s, " \t\r\n") {
		return StringSubject{}
	}
	s = strings.Replace(s, ":", ".", 1)
	i := strings.Index(s, ".")
	j := i + 1
	if i < 0 {
		i = len(s)
		j = i
	}
	return StringSubject{s: s, i: uint16(i), j: uint16(j)}
}

func FromParts(domain, typ string) StringSubject {
	b := strings.Builder{}
	b.Grow(len(domain) + 1 + len(typ))
	b.WriteString(domain)
	if typ != "" {
		b.WriteString(".")
		b.WriteString(typ)
	}
	return FromStr(b.String())
}

func (s StringSubject) Domain() string  { return s.s[:s.i] }
func (s StringSubject) Type() string    { return s.s[s.j:] }
func (s StringSubject) Subject() string { return s.s }
func (s StringSubject) String() string  { return s.Subject() }
func (s StringSubject) Parts() []string { return strings.Split(s.String(), ".") }
func (s StringSubject) Match(m string) bool {
	m = strings.Replace(m, ".", "\\.", -1)
	m = strings.Replace(m, "*", "[^\\.]+", -1)
	matched, _ := regexp.MatchString(`(?i)^`+m+`$`, s.String())
	return matched
}
