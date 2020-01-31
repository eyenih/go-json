package json

import (
	"io"
	"testing"

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

func (m *includedMapper) MapperFor(k string) Mapper {
	return nil
}

func TestParsingGoodGrammar(t *testing.T) {
	content := `
	{
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

	err := Parse(it, fsm)
	require.NoError(t, err)

	assert.Equal(t, Nil, fsm.s)

	require.Len(t, m.mappersFor, 2)
	assert.Equal(t, "data", m.mappersFor[0])
	assert.Equal(t, "included", m.mappersFor[1])
}

func BenchmarkParse(b *testing.B) {
	content := `
	{
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

	b.ReportAllocs()
	var err error
	for i := 0; i <= b.N; i++ {
		err = Parse(it, fsm)
	}

	if err != nil {
		panic("wtf?")
	}
}
