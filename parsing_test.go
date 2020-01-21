package json

import (
	"fmt"
	"io"
	_ "strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reader struct {
	buf []byte
}

func (r *reader) Read(p []byte) (int, error) {
	if 0 == len(r.buf) {
		return 0, io.EOF
	}
	n := copy(p[:1], r.buf[:1])
	r.buf = append([]byte{}, r.buf[1:]...)
	return n, nil
}

type objectMapper struct {
	keys []string
}

func (m *objectMapper) Key(k string) {
	m.keys = append(m.keys, k)
}

func TestParsingGoodGrammar(t *testing.T) {
	/*
		t.Run("only whitespaces", func(t *testing.T) {
			raws := map[string]rune{
				"space":          ' ',
				"new line":       '\n',
				"horizontal tab": '\t',
				"carrige return": '\r',
			}
			for _, v := range raws {
				str := strings.Repeat(string(v), 5)
				it := NewTextIterator(&reader{buf: []byte(str)})
				fsm := NewGrammarStateMachine()

				err := Parse(it, fsm)
				require.NoError(t, err)
			}
		})

		t.Run("null", func(t *testing.T) {
			str := "null"
			it := NewTextIterator(&reader{buf: []byte(str)})
			fsm := NewGrammarStateMachine()

			err := Parse(it, fsm)
			require.NoError(t, err)
		})
	*/

	content := `{
		"data": {
			"id": "test-id",
			"type": "test-type",
			"attributes": {
				"name": "test-name"
			},
			"relationships": {}
		},
		"included": [{
			"id": "test-id",
			"type": "test-type-included",
			"attributes": {
				"name": "test-name-included"
			}
		}]
	}`

	it := NewTextIterator(&reader{buf: []byte(content)})
	m := &objectMapper{}
	fsm := NewGrammarStateMachine(m)
	l := NewTestLogger()
	p := NewParser(l)

	err := p.Parse(it, fsm)
	require.NoError(t, err)
	require.Len(t, m.keys, 2)
	assert.Equal(t, "data", m.keys[0])
	assert.Equal(t, "included", m.keys[1])
}

type TestLogger struct {
}

func (l TestLogger) Trace(msg string, variables map[string]interface{}) {
	fmt.Println(msg)
	for k, v := range variables {
		switch a := v.(type) {
		case byte:
			if unicode.IsSpace(rune(a)) {
				fmt.Println(k, ":", "whitespace=", a)
			} else {
				fmt.Println(k, ":", string(a))
			}
		default:
			fmt.Println(k, ":", a)
		}
	}
}
func (l TestLogger) Debug(msg string, v map[string]interface{}) {
}
func (l TestLogger) Info(string, map[string]interface{})  {}
func (l TestLogger) Warn(string, map[string]interface{})  {}
func (l TestLogger) Error(string, map[string]interface{}) {}
func (l TestLogger) Fatal(string, map[string]interface{}) {}

func NewTestLogger() *TestLogger {
	return &TestLogger{}
}
