package subject_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/jrgensen/stream"
	"github.com/jrgensen/stream/subject"
)

func TestSubject(t *testing.T) {
	for _, test := range []struct {
		name string
		dom  string
		typ  string
	}{
		{
			name: "domain only",
			dom:  "domain",
		},
		{
			name: "domain and type",
			dom:  "domain",
			typ:  "type",
		},
		{
			name: "empty",
		},
	} {
		t.Run("FromStr:"+test.name, func(t *testing.T) {
			both := test.dom
			if test.typ != "" {
				both += "." + test.typ
			}
			s := subject.FromStr(both)
			if both != s.String() {
				t.Fatalf("exp '%+v' got '%+v'", both, s)
			}
			if test.dom != s.Domain() {
				t.Fatalf("exp domain '%v' got '%v'", test.dom, s.Domain())
			}
			if test.typ != s.Type() {
				t.Fatalf("exp type '%v' got '%v'", test.typ, s.Type())
			}
		})
		t.Run("FromParts:"+test.name, func(t *testing.T) {
			both := test.dom
			if test.typ != "" {
				both += "." + test.typ
			}
			s := subject.FromParts(test.dom, test.typ)
			if both != s.Subject() {
				t.Fatalf("exp '%+v' got '%+v'", both, s)
			}
			if test.dom != s.Domain() {
				t.Fatalf("exp domain '%v' got '%v'", test.dom, s.Domain())
			}
			if test.typ != s.Type() {
				t.Fatalf("exp type '%v' got '%v'", test.typ, s.Type())
			}
		})
	}
}

func BenchmarkSubjectLookup(b *testing.B) {
	s1 := subject.FromStr("foo:bar")
	s2 := subject.FromStr("foo:bar")
	m := make(map[string]struct{})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m[s1.Subject()] = struct{}{}
		if _, exist := m[s2.Subject()]; !exist {
			b.Fatalf("exp s2 exist")
		}
	}
}

func BenchmarkSubjectInterface(b *testing.B) {
	b.ReportAllocs()
	type Subject interface {
		Domain() string
		Type() string
		Subject() string
	}

	type T struct {
		A subject.StringSubject
		B Subject
	}

	b.Run("interface", func(b *testing.B) {
		msg := T{
			B: subject.FromStr("foo:bar"),
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg.B.Subject()
		}
	})

	b.Run("type", func(b *testing.B) {
		msg := T{
			A: subject.FromStr("foo:bar"),
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg.A.Subject()
		}
	})
}

func BenchmarkSubjectLookupString(b *testing.B) {
	s1 := StringSubject("foo:bar")
	s2 := StringSubject("foo:bar")
	m := make(map[StringSubject]struct{})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m[s1] = struct{}{}
		if _, exist := m[s2]; !exist {
			b.Fatalf("exp s2 exist")
		}
	}
}

func BenchmarkSubjectFromStr(b *testing.B) {
	var subj stream.Subject

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := subject.FromStr("foo:bar")
		subj = s
	}
	_ = subj
}

func BenchmarkSubjectFromParts(b *testing.B) {
	var subj stream.Subject
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		subj = subject.FromParts("a", "b")
	}
	_ = subj
}

func BenchmarkSubjectType(b *testing.B) {
	var typ string
	subj := subject.FromStr("foo:bar" + strconv.Itoa(b.N))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := subj.Type()
		typ = s

	}
	_ = typ
}

type StringSubject string

func (s StringSubject) Domain() string {
	return string(s)
}

func (s StringSubject) Type() string {
	i := strings.Index(string(s), ":")
	if i == -1 {
		i = len(s) - 1
	}
	return string(s[i+1])
}

func BenchmarkSubjectString(b *testing.B) {
	var subj StringSubject

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := StringSubject("foo:bar")
		subj = s
	}
	_ = subj
}

func BenchmarkSubjectStringType(b *testing.B) {
	var typ string
	subj := StringSubject("foo:bar")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typ = subj.Type()
	}
	_ = typ
}
