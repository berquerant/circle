package circle

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

func (*nothing) IsNothing() bool                     { return true }
func (*nothing) Get() (interface{}, bool)            { return nil, false }
func (*nothing) GetOrElse(v interface{}) interface{} { return v }
func (*nothing) OrElse(v Maybe) Maybe                { return v }
func (*nothing) Map(Mapper) Maybe                    { return nothingEntity }
func (*nothing) Filter(Filter) Maybe                 { return nothingEntity }
