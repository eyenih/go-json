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

type objectMapper struct {
	mappersFor []string
}

func (m *objectMapper) SetString(k, v string) {}

func (m *objectMapper) MapperFor(k string) Mapper {
	m.mappersFor = append(m.mappersFor, k)
	switch k {
	case "data":
		return &resourceMapper{}
	case "included":
		return &includedMapper{}
	}
	return nil
}

type resourceMapper struct {
}

func (m *resourceMapper) SetString(k, v string) {}

func (m *resourceMapper) MapperFor(k string) Mapper {
	switch k {
	case "attributes":
		return m
	case "relationships":
		return m
	}
	return nil
}

type includedMapper struct {
}

func (m *includedMapper) SetString(k, v string) {}

func (m *includedMapper) MapperFor(k string) Mapper {
	return nil
}

type basicObject struct {
	name   string
	number int
}

type basicObjectMapper struct {
	bo basicObject
}

func (m *basicObjectMapper) SetString(k, v string) {
	if k == "name" {
		m.bo.name = v
	}
}

func (m *basicObjectMapper) MapperFor(k string) Mapper {
	return nil
}

func TestParsingGoodGrammar(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		name := gofakeit.Name()
		number := gofakeit.Float32()
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
