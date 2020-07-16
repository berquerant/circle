package circle

type (
	// Executor provides an interface for applying function to iterator.
	Executor interface {
		Execute() (*Iterator, error)
	}
)

type (
	mapExecutor struct {
		f  Mapper
		it *Iterator
	}
)

// NewMapExecutor returns a new Executor for map.
//
// If f returns error, the argument of f is ignored, this does not yield it.
func NewMapExecutor(f Mapper, it *Iterator) Executor {
	return &mapExecutor{
		f:  f,
		it: it,
	}
}

func (s *mapExecutor) Execute() (*Iterator, error) {
	var f func() (interface{}, error)
	f = func() (interface{}, error) {
		x, err := s.it.Next()
		if err != nil {
			return nil, err
		}
		v, err := s.f.Apply(x)
		if err != nil {
			// ignore this value
			return f()
		}
		return v, nil
	}
	return NewIterator(f)
}
