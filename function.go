package circle

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/berquerant/circle/internal/reflection"
)

var (
	ErrInvalidMapper = errors.New("invalid mapper")
)

type (
	// Mapper is a func(A) (B, error) or func(A) B.
	Mapper interface {
		Apply(v interface{}) (interface{}, error)
	}

	mapper struct {
		f interface{}
	}
)

func isMapper(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 1) {
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

func (s *mapper) Apply(v interface{}) (ret interface{}, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	av, err := reflection.Convert(v, reflect.TypeOf(s.f).In(0), true)
	if err != nil {
		return nil, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{av})
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

var (
	ErrInvalidFilter = errors.New("invalid filter")
)

type (
	// Filter is a func(A) (bool, error) or func(A) bool.
	Filter interface {
		Apply(v interface{}) (bool, error)
	}

	filter struct {
		f interface{}
	}
)

func isFilter(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 1) {
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

func (s *filter) Apply(v interface{}) (ret bool, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = false
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	av, err := reflection.Convert(v, reflect.TypeOf(s.f).In(0), true)
	if err != nil {
		return false, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{av})
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

var (
	ErrInvalidAggregator = errors.New("invalid aggregator")
)

type (
	// Aggregator is a func(A, B) (B, error) or func(B, A) (B, error).
	Aggregator interface {
		Apply(x, y interface{}) (interface{}, error)
		Type() AggregatorType
	}

	AggregatorType int

	aggregator struct {
		f interface{}
		t AggregatorType
	}
)

const (
	UnknownAggregatorType AggregatorType = iota
	// RightAggregatorType indicates func(A, B) (B, error) or func(A, B) B.
	RightAggregatorType
	// LeftAggregatorType indicates func(B, A) (B, error) or func(B, A) B.
	LeftAggregatorType
	// PerfectAggregatorType indicates func(A, A) (A, error) or func(A, A) A.
	PerfectAggregatorType
)

func isRightAggregator(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 2) {
		return false
	}
	switch t.NumOut() {
	case 1:
		return t.In(1).String() == t.Out(0).String()
	case 2:
		return t.In(1).String() == t.Out(0).String() && t.Out(1).String() == "error"
	default:
		return false
	}
}

func isLeftAggregator(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 2) {
		return false
	}
	switch t.NumOut() {
	case 1:
		return t.In(0).String() == t.Out(0).String()
	case 2:
		return t.In(0).String() == t.Out(0).String() && t.Out(1).String() == "error"
	default:
		return false
	}
}

func getAggregatorType(f interface{}) AggregatorType {
	if isRightAggregator(f) {
		if isLeftAggregator(f) {
			return PerfectAggregatorType
		}
		return RightAggregatorType
	}
	if isLeftAggregator(f) {
		return LeftAggregatorType
	}
	return UnknownAggregatorType
}

// NewAggregator returns a new Aggregator.
// If f is not appropriate for Aggregator, returns ErrInvalidAggregator.
func NewAggregator(f interface{}) (Aggregator, error) {
	t := getAggregatorType(f)
	if t == UnknownAggregatorType {
		return nil, ErrInvalidAggregator
	}
	return &aggregator{
		f: f,
		t: t,
	}, nil
}

func (s *aggregator) Type() AggregatorType { return s.t }

func (s *aggregator) Apply(x, y interface{}) (ret interface{}, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	t := reflect.TypeOf(s.f)
	vx, err := reflection.Convert(x, t.In(0), true)
	if err != nil {
		return nil, err
	}
	vy, err := reflection.Convert(y, t.In(1), true)
	if err != nil {
		return nil, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{vx, vy})
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

var (
	ErrInvalidComparator = errors.New("invalid comparator")
)

type (
	// Comparator is a func(A, A) (bool, error) or func(A, A) bool.
	Comparator interface {
		Apply(x, y interface{}) (bool, error)
	}

	comparator struct {
		f interface{}
	}
)

func isComparator(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 2 &&
		t.In(0).String() == t.In(1).String()) {
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

// NewComparator returns a new Comparator.
// If f is not appropriate for Comparator, retrurns ErrInvalidComparator.
func NewComparator(f interface{}) (Comparator, error) {
	if !isComparator(f) {
		return nil, ErrInvalidComparator
	}
	return &comparator{
		f: f,
	}, nil
}

func (s *comparator) Apply(x, y interface{}) (ret bool, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = false
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	t := reflect.TypeOf(s.f)
	vx, err := reflection.Convert(x, t.In(0), true)
	if err != nil {
		return false, err
	}
	vy, err := reflection.Convert(y, t.In(1), true)
	if err != nil {
		return false, err
	}
	var (
		r  = reflect.ValueOf(s.f).Call([]reflect.Value{vx, vy})
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

var (
	ErrInvalidConsumer = errors.New("invalid consumer")
)

type (
	// Consumer is a func(A) error or func(A).
	Consumer interface {
		Apply(x interface{}) error
	}
	consumer struct {
		f interface{}
	}
)

func isConsumer(f interface{}) bool {
	t := reflect.TypeOf(f)
	if !(t.Kind() == reflect.Func && t.NumIn() == 1) {
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

// NewConsumer returns a new Consumer.
func NewConsumer(f interface{}) (Consumer, error) {
	if !isConsumer(f) {
		return nil, ErrInvalidConsumer
	}
	return &consumer{
		f: f,
	}, nil
}

func (s *consumer) Apply(x interface{}) (rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rerr = fmt.Errorf("%w %s", ErrApply, err)
		}
	}()
	t := reflect.TypeOf(s.f)
	vx, err := reflection.Convert(x, t.In(0), true)
	if err != nil {
		return err
	}
	var (
		r = reflect.ValueOf(s.f).Call([]reflect.Value{vx})
	)
	if len(r) == 1 {
		r0 := r[0].Interface()
		if err, ok := r0.(error); ok {
			return err
		}
	}
	return nil
}
