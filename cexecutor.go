package circle

type (
	// ConsumeExecutor provides an interface for applying consumer function to iterator.
	ConsumeExecutor interface {
		ConsumeExecute() error
	}
)

type (
	consumeExecutor struct {
		f  Consumer
		it Iterator
	}
)

// NewConsumeExecutor returns a new ConsumeExecutor.
func NewConsumeExecutor(f Consumer, it Iterator) ConsumeExecutor {
	return &consumeExecutor{
		f:  f,
		it: it,
	}
}

func (s *consumeExecutor) ConsumeExecute() error {
	for {
		x, err := s.it.Next()
		if err == ErrEOI {
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.f.Apply(x); err != nil {
			return err
		}
	}
}
