package circle

import (
	"circle/internal/reflection"
	"errors"
	"reflect"
)

var (
	ErrApply = errors.New("apply error")
)

type (
	maybeMapper struct {
		f Mapper
	}
)

// NewMaybeMapper returns a new Mapper for Maybe.
//
// If f returns error or argument is nothing, returns nothing.
func NewMaybeMapper(f interface{}) (Mapper, error) {
	m, err := NewMapper(f)
	if err != nil {
		return nil, err
	}
	return &maybeMapper{f: m}, nil
}

func (s *maybeMapper) Apply(v interface{}) (interface{}, error) {
	x, ok := v.(Maybe)
	if !ok {
		return nil, ErrApply
	}
	return x.Map(s.f), nil
}

type (
	eitherMapper struct {
		f Mapper
	}
)

// NewEitherMapper returns a new Mapper for Either.
//
// If f returns error or argument is left, returns left.
func NewEitherMapper(f interface{}) (Mapper, error) {
	m, err := NewMapper(f)
	if err != nil {
		return nil, err
	}
	return &eitherMapper{f: m}, nil
}

func (s *eitherMapper) Apply(v interface{}) (interface{}, error) {
	x, ok := v.(Either)
	if !ok {
		return nil, ErrApply
	}
	return x.Map(s.f), nil
}

type (
	tupleMapper struct {
		f interface{}
	}
)

// NewTupleMapper returns a new Mapper for Tuple.
//
// If argument is not Tuple or number of parameters of f is not equal to size of argument Tuple, returns error.
func NewTupleMapper(f interface{}) (Mapper, error) {
	if !isTupleMapper(f) {
		return nil, ErrInvalidMapper
	}
	return &tupleMapper{
		f: f,
	}, nil
}

func isTupleMapper(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumOut() == 2 &&
		t.Out(1).String() == "error"
}

func (s *tupleMapper) Apply(v interface{}) (interface{}, error) {
	x, ok := v.(Tuple)
	if !ok {
		return nil, ErrApply
	}
	t := reflect.TypeOf(s.f)
	if x.Size() != t.NumIn() {
		return nil, ErrApply
	}
	a := make([]reflect.Value, x.Size())
	for i := 0; i < x.Size(); i++ {
		p, ok := x.Get(i)
		if !ok {
			return nil, ErrApply
		}
		v, err := reflection.Convert(p, t.In(i), true)
		if err != nil {
			return nil, err
		}
		a[i] = v
	}
	var (
		r  = reflect.ValueOf(s.f).Call(a)
		r0 = r[0].Interface()
		r1 = r[1].Interface()
	)
	if err, ok := r1.(error); ok {
		return r0, err
	}
	return r0, nil
}
