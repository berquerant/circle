package circle

import (
	"circle/internal/reflection"
	"errors"
	"reflect"
)

var (
	ErrInvalidMapper = errors.New("invalid mapper")
)

type (
	// Mapper is a func(A) (B, error)
	Mapper interface {
		Apply(v interface{}) (interface{}, error)
	}

	mapper struct {
		f interface{}
	}
)

func isMapper(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 1 && t.NumOut() == 2 &&
		t.Out(1).String() == "error"
}

// NewMapper returns a new Mapper.
// If f is not appropriate for Mapper, returns ErrInvalidMapper.
func NewMapper(f interface{}) (Mapper, error) {
	if !isMapper(f) {
		return nil, ErrInvalidMapper
	}
	return &mapper{
		f: f,
	}, nil
}

func (s *mapper) Apply(v interface{}) (interface{}, error) {
	av, err := reflection.Convert(v, reflect.TypeOf(s.f).In(0), true)
	if err != nil {
		return nil, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{av})
		r0 = r[0].Interface()
		r1 = r[1].Interface()
	)
	if err, ok := r1.(error); ok {
		return r0, err
	}
	return r0, nil
}

var (
	ErrInvalidFilter = errors.New("invalid filter")
)

type (
	// Filter is a func(A) (bool, error)
	Filter interface {
		Apply(v interface{}) (bool, error)
	}

	filter struct {
		f interface{}
	}
)

func isFilter(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 1 && t.NumOut() == 2 &&
		t.Out(0).Kind() == reflect.Bool && t.Out(1).String() == "error"
}

// NewFilter returns a new Filter.
// If f is not appropriate for Filter, returns ErrInvalidFilter.
func NewFilter(f interface{}) (Filter, error) {
	if !isFilter(f) {
		return nil, ErrInvalidFilter
	}
	return &filter{
		f: f,
	}, nil
}

func (s *filter) Apply(v interface{}) (bool, error) {
	av, err := reflection.Convert(v, reflect.TypeOf(s.f).In(0), true)
	if err != nil {
		return false, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{av})
		r0 = r[0].Bool()
		r1 = r[1].Interface()
	)
	if err, ok := r1.(error); ok {
		return r0, err
	}
	return r0, nil
}
