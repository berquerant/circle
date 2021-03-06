package circle

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/berquerant/circle/internal/reflection"
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
// If you want to convert Maybe[A] to B, f is a func(A) (B, error) or func(A) B.
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
// If you want to convert Either[_, A] to B, f is a func(A) (B, error) or func(A) B.
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
// If you want to convert Tuple(A1, A2, ..., An), f is a func(A1, A2, ..., An) (B, error) or func(A1, A2, ..., An) B.
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
	if t.Kind() != reflect.Func {
		return false
	}
	switch t.NumOut() {
	case 1:
		return true
	case 2:
		return t.Out(1).String() == "error"
	default:
		return false
	}
}

func (s *tupleMapper) Apply(v interface{}) (ret interface{}, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
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
	)
	if len(r) == 2 {
		r1 := r[1].Interface()
		if err, ok := r1.(error); ok {
			return r0, err
		}
	}
	return r0, nil
}

type (
	tupleFilter struct {
		f interface{}
	}
)

// NewTupleFilter returns a new Filter for Tuple.
//
// If you want to filter Tuple(A1, A2, ..., An), f is a func(A1, A2, ..., An) (bool, error) or func(A1, A2, ..., An) bool.
//
// If argument is not Tuple or number of parameters of f is not equal to size of argument Tuple, returns error.
func NewTupleFilter(f interface{}) (Filter, error) {
	if !isTupleFilter(f) {
		return nil, ErrInvalidFilter
	}
	return &tupleFilter{
		f: f,
	}, nil
}

func isTupleFilter(f interface{}) bool {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return false
	}
	switch t.NumOut() {
	case 1:
		return t.Out(0).Kind() == reflect.Bool
	case 2:
		return t.Out(0).Kind() == reflect.Bool && t.Out(1).String() == "error"
	default:
		return false
	}
}

func (s *tupleFilter) Apply(v interface{}) (ret bool, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = false
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	x, ok := v.(Tuple)
	if !ok {
		return false, ErrApply
	}
	t := reflect.TypeOf(s.f)
	if x.Size() != t.NumIn() {
		return false, ErrApply
	}
	a := make([]reflect.Value, x.Size())
	for i := 0; i < x.Size(); i++ {
		p, ok := x.Get(i)
		if !ok {
			return false, ErrApply
		}
		v, err := reflection.Convert(p, t.In(i), true)
		if err != nil {
			return false, err
		}
		a[i] = v
	}
	var (
		r  = reflect.ValueOf(s.f).Call(a)
		r0 = r[0].Bool()
	)
	if len(r) == 2 {
		r1 := r[1].Interface()
		if err, ok := r1.(error); ok {
			return r0, err
		}
	}
	return r0, nil
}

type (
	tupleConsumer struct {
		f interface{}
	}
)

// NewTupleConsumer returns a new Consumer for Tuple.
//
// If you want to consume Tuple(A1, A2, ..., An), f is a func(A1, A2, ..., An) error or func(A1, A2, ..., An).
//
// If argument is not Tuple or number of parameters of f is not equal to size of argument Tuple, returns error.
func NewTupleConsumer(f interface{}) (Consumer, error) {
	if !isTupleConsumer(f) {
		return nil, ErrInvalidConsumer
	}
	return &tupleConsumer{
		f: f,
	}, nil
}

func isTupleConsumer(f interface{}) bool {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return false
	}
	switch t.NumOut() {
	case 0:
		return true
	case 1:
		return t.Out(0).String() == "error"
	default:
		return false
	}
}

func (s *tupleConsumer) Apply(v interface{}) (rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	x, ok := v.(Tuple)
	if !ok {
		return ErrApply
	}
	t := reflect.TypeOf(s.f)
	if x.Size() != t.NumIn() {
		return ErrApply
	}
	a := make([]reflect.Value, x.Size())
	for i := 0; i < x.Size(); i++ {
		p, ok := x.Get(i)
		if !ok {
			return ErrApply
		}
		v, err := reflection.Convert(p, t.In(i), true)
		if err != nil {
			return err
		}
		a[i] = v
	}
	var (
		r = reflect.ValueOf(s.f).Call(a)
	)
	if len(r) == 1 {
		r0 := r[0].Interface()
		if err, ok := r0.(error); ok {
			return err
		}
	}
	return nil
}

type (
	maybeConsumer struct {
		fj Consumer
		fn Consumer
	}
)

// NewMaybeConsumer returns a new Consumer for Maybe.
// If you want to consume Maybe[A] that is not nothing, f is a func(A) error.
// g is a func() error to consume Nothing.
func NewMaybeConsumer(f interface{}, g func() error) (Consumer, error) {
	fj, err := NewConsumer(f)
	if err != nil {
		return nil, err
	}
	fn, err := NewConsumer(func(interface{}) error { return g() })
	if err != nil {
		return nil, err
	}
	return &maybeConsumer{
		fj: fj,
		fn: fn,
	}, nil
}

func (s *maybeConsumer) Apply(x interface{}) error {
	v, ok := x.(Maybe)
	if !ok {
		return ErrApply
	}
	return v.Consume(s.fj, s.fn)
}

type (
	eitherConsumer struct {
		fr Consumer
		fl Consumer
	}
)

// NewEitherConsumer returns a new Consumer for Either.
// If you want to consume Either[A, B],
// f is a func(A) error, g is a func(B) error.
func NewEitherConsumer(f, g interface{}) (Consumer, error) {
	fl, err := NewConsumer(f)
	if err != nil {
		return nil, err
	}
	fr, err := NewConsumer(g)
	if err != nil {
		return nil, err
	}
	return &eitherConsumer{
		fl: fl,
		fr: fr,
	}, nil
}

func (s *eitherConsumer) Apply(x interface{}) error {
	v, ok := x.(Either)
	if !ok {
		return ErrApply
	}
	return v.Consume(s.fl, s.fr)
}
