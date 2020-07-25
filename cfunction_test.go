package circle_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/berquerant/circle"

	"github.com/stretchr/testify/assert"
)

type (
	testcaseMaybeMapper struct {
		title        string
		arg          interface{}
		f            func(int) (int, error)
		want         circle.Maybe
		isApplyError bool
	}
)

func (s *testcaseMaybeMapper) test(t *testing.T) {
	f, err := circle.NewMaybeMapper(s.f)
	assert.Nil(t, err)
	v, err := f.Apply(s.arg)
	assert.Equal(t, s.isApplyError, err != nil)
	if s.isApplyError {
		return
	}
	got, ok := v.(circle.Maybe)
	if !assert.True(t, ok) {
		return
	}
	gotVal, gotOK := got.Get()
	wantVal, wantOK := s.want.Get()
	assert.Equal(t, wantOK, gotOK)
	assert.Equal(t, wantVal, gotVal)
}

func TestMaybeMapper(t *testing.T) {
	for _, tc := range []*testcaseMaybeMapper{
		{
			title:        "not maybe",
			arg:          1,
			f:            func(int) (int, error) { return 0, nil },
			isApplyError: true,
		},
		{
			title: "nothing",
			arg:   circle.NewNothing(),
			f:     func(int) (int, error) { return 0, nil },
			want:  circle.NewNothing(),
		},
		{
			title: "ok",
			arg:   circle.NewJust(1),
			f:     func(x int) (int, error) { return x + 1, nil },
			want:  circle.NewJust(2),
		},
		{
			title: "failure",
			arg:   circle.NewJust(1),
			f:     func(int) (int, error) { return 0, errors.New("failure") },
			want:  circle.NewNothing(),
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseMaybeConsumer struct {
		title        string
		arg          interface{}
		fg           func(chan<- interface{}) (interface{}, func() error)
		want         interface{}
		isApplyError bool
	}
)

func (s *testcaseMaybeConsumer) test(t *testing.T) {
	ch := make(chan interface{})
	f, g := s.fg(ch)
	c, err := circle.NewMaybeConsumer(f, g)
	assert.Nil(t, err)
	go func() {
		assert.Equal(t, s.isApplyError, c.Apply(s.arg) != nil)
		close(ch)
	}()
	got, ok := <-ch
	assert.Equal(t, ok, s.want != nil)
	if ok && s.want != nil {
		assert.Equal(t, s.want, got)
	}
}

func TestMaybeConsumer(t *testing.T) {
	for _, tc := range []*testcaseMaybeConsumer{
		{
			title: "not maybe",
			arg:   1,
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v int) error {
						ch <- v
						return nil
					}, func() error {
						ch <- -1
						return nil
					}
			},
			isApplyError: true,
		},
		{
			title: "nothing",
			arg:   circle.NewNothing(),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v int) error {
						ch <- v
						return nil
					}, func() error {
						ch <- -1
						return nil
					}
			},
			want: -1,
		},
		{
			title: "just",
			arg:   circle.NewJust(100),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v int) error {
						ch <- v
						return nil
					}, func() error {
						ch <- -1
						return nil
					}
			},
			want: 100,
		},
		{
			title: "just error",
			arg:   circle.NewJust(100),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v int) error {
						return errors.New("just")
					}, func() error {
						ch <- -1
						return nil
					}
			},
			isApplyError: true,
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseEitherConsumer struct {
		title        string
		arg          interface{}
		fg           func(chan<- interface{}) (interface{}, interface{})
		want         interface{}
		isApplyError bool
	}
)

func (s *testcaseEitherConsumer) test(t *testing.T) {
	ch := make(chan interface{})
	f, g := s.fg(ch)
	c, err := circle.NewEitherConsumer(f, g)
	assert.Nil(t, err)
	go func() {
		assert.Equal(t, s.isApplyError, c.Apply(s.arg) != nil)
		close(ch)
	}()
	got, ok := <-ch
	assert.Equal(t, ok, s.want != nil)
	if ok && s.want != nil {
		assert.Equal(t, s.want, got)
	}
}

func TestEitherConsumer(t *testing.T) {
	for _, tc := range []*testcaseEitherConsumer{
		{
			title: "not either",
			arg:   1,
			fg: func(ch chan<- interface{}) (interface{}, interface{}) {
				return func(x int) error {
						ch <- x
						return nil
					}, func(x int) error {
						ch <- x
						return nil
					}
			},
			isApplyError: true,
		},
		{
			title: "left",
			arg:   circle.NewLeft(1),
			fg: func(ch chan<- interface{}) (interface{}, interface{}) {
				return func(x int) error {
						ch <- x
						return nil
					}, func(x int) error {
						ch <- x + 1
						return nil
					}
			},
			want: 1,
		},
		{
			title: "right",
			arg:   circle.NewRight(1),
			fg: func(ch chan<- interface{}) (interface{}, interface{}) {
				return func(x int) error {
						ch <- x + 1
						return nil
					}, func(x int) error {
						ch <- x
						return nil
					}
			},
			want: 1,
		},
		{
			title: "right error",
			arg:   circle.NewRight(1),
			fg: func(ch chan<- interface{}) (interface{}, interface{}) {
				return func(x int) error {
						ch <- x + 1
						return nil
					}, func(x int) error {
						return errors.New("err")
					}
			},
			isApplyError: true,
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseEitherMapper struct {
		title        string
		arg          interface{}
		f            func(int) (int, error)
		want         circle.Either
		isApplyError bool
	}
)

func (s *testcaseEitherMapper) test(t *testing.T) {
	f, err := circle.NewEitherMapper(s.f)
	assert.Nil(t, err)
	v, err := f.Apply(s.arg)
	assert.Equal(t, s.isApplyError, err != nil)
	if s.isApplyError {
		return
	}
	got, ok := v.(circle.Either)
	if !assert.True(t, ok) {
		return
	}
	{
		gotVal, gotOK := got.Left()
		wantVal, wantOK := s.want.Left()
		assert.Equal(t, wantOK, gotOK)
		assert.Equal(t, wantVal, gotVal)
	}
	{
		gotVal, gotOK := got.Right()
		wantVal, wantOK := s.want.Right()
		assert.Equal(t, wantOK, gotOK)
		assert.Equal(t, wantVal, gotVal)
	}
}

func TestEitherMapper(t *testing.T) {
	for _, tc := range []*testcaseEitherMapper{
		{
			title:        "not either",
			arg:          1,
			f:            func(int) (int, error) { return 0, nil },
			isApplyError: true,
		},
		{
			title: "left",
			arg:   circle.NewLeft("error"),
			f:     func(int) (int, error) { return 0, nil },
			want:  circle.NewLeft("error"),
		},
		{
			title: "right",
			arg:   circle.NewRight(1),
			f:     func(x int) (int, error) { return x + 1, nil },
			want:  circle.NewRight(2),
		},
		{
			title: "right",
			arg:   circle.NewRight(1),
			f:     func(int) (int, error) { return 0, errors.New("error") },
			want:  circle.NewLeft(errors.New("error")),
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseTupleMapper struct {
		title        string
		arg          interface{}
		f            interface{}
		want         interface{}
		isApplyError bool
	}
)

func (s *testcaseTupleMapper) test(t *testing.T) {
	f, err := circle.NewTupleMapper(s.f)
	assert.Nil(t, err)
	got, err := f.Apply(s.arg)
	assert.Equal(t, s.isApplyError, err != nil)
	if s.isApplyError {
		return
	}
	assert.Equal(t, s.want, got)
}

func TestTupleMapper(t *testing.T) {
	for _, tc := range []*testcaseTupleMapper{
		{
			title:        "not tuple",
			arg:          1,
			f:            func(int, string) (int, error) { return 0, nil },
			isApplyError: true,
		},
		{
			title:        "invalid argument",
			arg:          circle.NewTuple("failure", 100),
			f:            func(int, int) (int, error) { return 0, nil },
			isApplyError: true,
		},
		{
			title:        "tuple size error",
			arg:          circle.NewTuple(1),
			f:            func(int, string) (int, error) { return 0, nil },
			isApplyError: true,
		},
		{
			title: "empty tuple",
			arg:   circle.NewTuple(),
			f:     func() (int, error) { return 1, nil },
			want:  1,
		},
		{
			title: "unit",
			arg:   circle.NewTuple(1),
			f:     func(x int) (int, error) { return x + 1, nil },
			want:  2,
		},
		{
			title: "tuple",
			arg:   circle.NewTuple(1, "two"),
			f:     func(x int, y string) (string, error) { return fmt.Sprintf("%d-%s", x, y), nil },
			want:  "1-two",
		},
		{
			title: "tuple tuple",
			arg:   circle.NewTuple(circle.NewTuple(1), circle.NewTuple(2)),
			f: func(x, y circle.Tuple) (int, error) {
				a, _ := x.Get(0)
				b, _ := y.Get(0)
				return a.(int) + b.(int), nil
			},
			want: 3,
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseTupleFilter struct {
		title        string
		arg          interface{}
		f            interface{}
		want         bool
		isApplyError bool
	}
)

func (s *testcaseTupleFilter) test(t *testing.T) {
	f, err := circle.NewTupleFilter(s.f)
	assert.Nil(t, err)
	got, err := f.Apply(s.arg)
	assert.Equal(t, s.isApplyError, err != nil)
	if s.isApplyError {
		return
	}
	assert.Equal(t, s.want, got)
}

func TestTupleFilter(t *testing.T) {
	for _, tc := range []*testcaseTupleFilter{
		{
			title:        "not tuple",
			arg:          1,
			f:            func(int, int) (bool, error) { return true, nil },
			isApplyError: true,
		},
		{
			title:        "invalid argument",
			arg:          circle.NewTuple(1, "2"),
			f:            func(int, int) (bool, error) { return true, nil },
			isApplyError: true,
		},
		{
			title:        "tuple size error",
			arg:          circle.NewTuple(1),
			f:            func(int, int) (bool, error) { return true, nil },
			isApplyError: true,
		},
		{
			title: "empty tuple",
			arg:   circle.NewTuple(),
			f:     func() (bool, error) { return true, nil },
			want:  true,
		},
		{
			title: "unit",
			arg:   circle.NewTuple(1),
			f:     func(x int) (bool, error) { return x == 1, nil },
			want:  true,
		},
		{
			title: "unit false",
			arg:   circle.NewTuple(1),
			f:     func(x int) (bool, error) { return x != 1, nil },
			want:  false,
		},
		{
			title: "tuple",
			arg:   circle.NewTuple(1, 2),
			f:     func(x, y int) (bool, error) { return x == 1 && y == 2, nil },
			want:  true,
		},
		{
			title: "tuple tuple",
			arg:   circle.NewTuple(circle.NewTuple(1), circle.NewTuple(2)),
			f: func(x, y circle.Tuple) (bool, error) {
				a, _ := x.Get(0)
				b, _ := y.Get(0)
				return a.(int) == 1 && b.(int) == 2, nil
			},
			want: true,
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
