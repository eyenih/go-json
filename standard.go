package json

import (
	"reflect"
)

type StandardMapper struct {
	instance interface{}
	keys     map[string]string
}

func NewStandardMapper(instance interface{}) *StandardMapper {
	return &StandardMapper{instance, map[string]string{}}
}

func (m *StandardMapper) Compile() {
	s := reflect.TypeOf(m.instance)

	v := s.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		m.keys[f.Tag.Get("json")] = f.Name
	}
}

func (m *StandardMapper) SetString(k, v string) {
}

func (m *StandardMapper) SetFloat(k string, v float64) {
}

func (m *StandardMapper) MapperFor(k string) Mapper {
	return nil
}
