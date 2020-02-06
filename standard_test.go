package json

import (
	"testing"

	_ "github.com/brianvoe/gofakeit/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandardMapper(t *testing.T) {
	t.Run("implement Mapper", func(t *testing.T) {
		m := &StandardMapper{}

		require.Implements(t, (*Mapper)(nil), m)
	})

	t.Run("compile", func(t *testing.T) {
		type X struct {
			A string `json:"aKey"`
			B int    `json:"bKey"`
		}

		m := NewStandardMapper(&X{})

		m.Compile()

		assert.Equal(t, "A", m.keys["aKey"])
		assert.Contains(t, "B", m.keys["bKey"])
	})
	//m.SetString(gofakeit.FirstName(), gofakeit.LastName())
	//m.SetFloat(string, float64)
	//m.MapperFor(string) Mapper
}
