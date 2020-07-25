package circle

import (
	"fmt"
	"strings"
)

type (
	// Maybe is an optional value.
	Maybe interface {
		// IsNothing returns true if this has no value.
		IsNothing() bool
		// Get returns the value of this.
		// If this is nothing, returns false.
		Get() (interface{}, bool)
		// GetOrElse returns the value of this if this is not nothing,
		// else returns v.
		GetOrElse(v interface{}) interface{}
		// OrElse returns this if this is not nothing,
		// else returns v.
		OrElse(v Maybe) Maybe
		// Map applies f to the value of this if this is not nothing.
		Map(f Mapper) Maybe
		// Filter applies f to the value of this if this is not nothing.
		Filter(f Filter) Maybe
		// Consume applies f to the value of this if this is not nothing,
		// else calls g.
		Consume(f, g Consumer) error
	}

	just struct {
		v interface{}
	}
	nothing struct {
	}
)

var (
	nothingEntity = &nothing{}
)

// NewJust returns a new Maybe that has value.
func NewJust(v interface{}) Maybe { return &just{v: v} }

// NewNothing returns a new Maybe tha has no value.
func NewNothing() Maybe { return nothingEntity }

func (*just) IsNothing() bool                       { return false }
func (s *just) Get() (interface{}, bool)            { return s.v, true }
func (s *just) GetOrElse(v interface{}) interface{} { return s.v }
func (s *just) OrElse(_ Maybe) Maybe                { return s }
func (s *just) Map(f Mapper) Maybe {
	v, err := f.Apply(s.v)
	if err != nil {
		return nothingEntity
	}
	return &just{v: v}
}
func (s *just) Filter(f Filter) Maybe {
	if ok, err := f.Apply(s.v); ok && err == nil {
		return s
	}
	return nothingEntity
}
func (s *just) Consume(f, _ Consumer) error { return f.Apply(s.v) }
func (s *just) String() string              { return fmt.Sprintf("Just(%v)", s.v) }

func (*nothing) IsNothing() bool                     { return true }
func (*nothing) Get() (interface{}, bool)            { return nil, false }
func (*nothing) GetOrElse(v interface{}) interface{} { return v }
func (*nothing) OrElse(v Maybe) Maybe                { return v }
func (*nothing) Map(Mapper) Maybe                    { return nothingEntity }
func (*nothing) Filter(Filter) Maybe                 { return nothingEntity }
func (*nothing) Consume(_, g Consumer) error         { return g.Apply(nothingEntity) }
func (*nothing) String() string                      { return "Nothing" }

type (
	// Either contains successful right or failed left value.
	Either interface {
		// IsLeft returns true if this has failed value.
		IsLeft() bool
		// IsRight returns true if this has successful value.
		IsRight() bool
		// Left returns left value.
		// If this is not left, returns false.
		Left() (interface{}, bool)
		// Right returns right value.
		// If this is not right, returns false.
		Right() (interface{}, bool)
		// GetOrElse returns right value if this is right else returns v.
		GetOrElse(v interface{}) interface{}
		// Map applies f to value if this is right.
		// If f returns error, returns left.
		Map(f Mapper) Either
		// ToMaybe converts this to Maybe.
		// If this is right, returns Just,
		// else returns Nothing.
		ToMaybe() Maybe
		// Consume applies g to this if this is right,
		// else f.
		Consume(f, g Consumer) error
	}

	left struct {
		v interface{}
	}
	right struct {
		v interface{}
	}
)

// NewRight returns a new Right.
func NewRight(v interface{}) Either { return &right{v: v} }

// NewLeft returns a new Left.
func NewLeft(v interface{}) Either { return &left{v: v} }

func (*left) IsLeft() bool                        { return true }
func (*left) IsRight() bool                       { return false }
func (s *left) Left() (interface{}, bool)         { return s.v, true }
func (s *left) Right() (interface{}, bool)        { return nil, false }
func (*left) GetOrElse(v interface{}) interface{} { return v }
func (s *left) Map(f Mapper) Either               { return s }
func (*left) ToMaybe() Maybe                      { return nothingEntity }
func (s *left) Consume(f, _ Consumer) error       { return f.Apply(s.v) }
func (s *left) String() string                    { return fmt.Sprintf("Left(%v)", s.v) }

func (*right) IsLeft() bool                        { return false }
func (*right) IsRight() bool                       { return true }
func (*right) Left() (interface{}, bool)           { return nil, false }
func (s *right) Right() (interface{}, bool)        { return s.v, true }
func (s *right) GetOrElse(interface{}) interface{} { return s.v }
func (s *right) Map(f Mapper) Either {
	v, err := f.Apply(s.v)
	if err != nil {
		return &left{v: err}
	}
	return &right{v: v}
}
func (s *right) ToMaybe() Maybe              { return &just{v: s.v} }
func (s *right) Consume(_, g Consumer) error { return g.Apply(s.v) }
func (s *right) String() string              { return fmt.Sprintf("Right(%v)", s.v) }

type (
	// Tuple is an immutable array.
	Tuple interface {
		// Size returns the size of this.
		Size() int
		// Get returns an element.
		// If i is out of range, returns false.
		Get(i int) (interface{}, bool)
	}

	tuple struct {
		v []interface{}
	}
)

// NewTuple returns a new Tuple.
func NewTuple(v ...interface{}) Tuple { return &tuple{v: v} }

func (s *tuple) Size() int { return len(s.v) }
func (s *tuple) Get(i int) (interface{}, bool) {
	if i < 0 || i >= len(s.v) {
		return nil, false
	}
	return s.v[i], true
}
func (s *tuple) String() string {
	a := make([]string, len(s.v))
	for i, x := range s.v {
		a[i] = fmt.Sprint(x)
	}
	return fmt.Sprintf("Tuple(%s)", strings.Join(a, ","))
}
