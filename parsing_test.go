package json

import (
	"fmt"
	"io"
	"testing"

	"github.com/brianvoe/gofakeit/v4"
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

type basicObject struct {
	name   string
	number float64
}

type basicObjectMapper struct {
	bo basicObject
}

func (m *basicObjectMapper) SetString(k, v string) {
	if k == "name" {
		m.bo.name = v
	}
}

func (m *basicObjectMapper) SetFloat(k string, v float64) {
	if k == "number" {
		m.bo.number = v
	}
}

func (m *basicObjectMapper) MapperFor(k string) Mapper {
	return nil
}

func TestParsingGoodGrammar(t *testing.T) {
	gofakeit.Seed(0)
	t.Run("object", func(t *testing.T) {
		name := gofakeit.Name()
		number := gofakeit.Float64()
		content := fmt.Sprintf(`{
			"name": "%s",
			"number": %.2f
		}`, name, number)

		it := NewTextIterator(&reader{buf: []byte(content)})
		m := &basicObjectMapper{}
		fsm := NewGrammarStateMachine(m)

		err := Parse(it, fsm)
		require.NoError(t, err)

		assert.Equal(t, name, m.bo.name)
		assert.Equal(t, number, m.bo.number)
	})
}

func BenchmarkParse(b *testing.B) {
	name := gofakeit.Name()
	number := gofakeit.Float32()
	content := fmt.Sprintf(`{
			"name": "%s",
			"number": %.2f
		}`, name, number)

	it := NewTextIterator(&reader{buf: []byte(content)})
	m := &basicObjectMapper{}
	fsm := NewGrammarStateMachine(m)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := Parse(it, fsm)

		if err != nil {
			b.Fatal(err)
		}
	}
}
