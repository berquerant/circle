package circle_test

import (
	"errors"
	"testing"

	"github.com/berquerant/circle"

	"github.com/stretchr/testify/assert"
)

type (
	testcaseMaybeMap struct {
		title string
		arg   circle.Maybe
		f     func(int) (int, error)
		want  circle.Maybe
	}
)

func (s *testcaseMaybeMap) test(t *testing.T) {
	f, err := circle.NewMapper(s.f)
	assert.Nil(t, err)
	gotVal, gotOK := s.arg.Map(f).Get()
	wantVal, wantOK := s.want.Get()
	assert.Equal(t, wantOK, gotOK)
	assert.Equal(t, wantVal, gotVal)
}

func TestMaybeMap(t *testing.T) {
	for _, tc := range []*testcaseMaybeMap{
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
		{
			title: "nothing",
			arg:   circle.NewNothing(),
			f:     func(int) (int, error) { return 0, nil },
			want:  circle.NewNothing(),
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseMaybeFilter struct {
		title string
		arg   circle.Maybe
		f     func(int) (bool, error)
		want  circle.Maybe
	}
)

func (s *testcaseMaybeFilter) test(t *testing.T) {
	f, err := circle.NewFilter(s.f)
	assert.Nil(t, err)
	gotVal, gotOK := s.arg.Filter(f).Get()
	wantVal, wantOK := s.want.Get()
	assert.Equal(t, wantOK, gotOK)
	assert.Equal(t, wantVal, gotVal)
}

func TestMaybeFilter(t *testing.T) {
	for _, tc := range []*testcaseMaybeFilter{
		{
			title: "ok",
			arg:   circle.NewJust(1),
			f:     func(int) (bool, error) { return true, nil },
			want:  circle.NewJust(1),
		},
		{
			title: "exclude",
			arg:   circle.NewJust(1),
			f:     func(int) (bool, error) { return false, nil },
			want:  circle.NewNothing(),
		},
		{
			title: "error",
			arg:   circle.NewJust(1),
			f:     func(int) (bool, error) { return false, errors.New("failure") },
			want:  circle.NewNothing(),
		},
		{
			title: "nothing ok",
			arg:   circle.NewNothing(),
			f:     func(int) (bool, error) { return true, nil },
			want:  circle.NewNothing(),
		},
		{
			title: "nothing exclude",
			arg:   circle.NewNothing(),
			f:     func(int) (bool, error) { return false, nil },
			want:  circle.NewNothing(),
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseMaybeConsume struct {
		title   string
		arg     circle.Maybe
		fg      func(chan<- interface{}) (interface{}, func() error)
		wantErr error
		wantVal interface{}
	}
)

func (s *testcaseMaybeConsume) test(t *testing.T) {
	ch := make(chan interface{})
	f, g := s.fg(ch)
	cf, err := circle.NewConsumer(f)
	assert.Nil(t, err)
	cg, err := circle.NewConsumer(func(interface{}) error { return g() })
	assert.Nil(t, err)
	go func() {
		assert.Equal(t, s.wantErr, s.arg.Consume(cf, cg))
		close(ch)
	}()
	gotVal, ok := <-ch
	assert.Equal(t, ok, s.wantVal != nil)
	if ok && s.wantVal != nil {
		assert.Equal(t, s.wantVal, gotVal)
	}
}

func TestMaybeConsume(t *testing.T) {
	for _, tc := range []*testcaseMaybeConsume{
		{
			title: "nothing error",
			arg:   circle.NewNothing(),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v interface{}) error {
						ch <- v
						return nil
					}, func() error {
						return errors.New("nth")
					}
			},
			wantErr: errors.New("nth"),
		},
		{
			title: "nothing value",
			arg:   circle.NewNothing(),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v interface{}) error {
						ch <- v
						return nil
					}, func() error {
						ch <- "got nothing"
						return nil
					}
			},
			wantVal: "got nothing",
		},
		{
			title: "just error",
			arg:   circle.NewJust("tea"),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v interface{}) error {
						return errors.New("just")
					}, func() error {
						ch <- "nth"
						return nil
					}
			},
			wantErr: errors.New("just"),
		},
		{
			title: "just value",
			arg:   circle.NewJust("tea"),
			fg: func(ch chan<- interface{}) (interface{}, func() error) {
				return func(v string) error {
						ch <- v
						return nil
					}, func() error {
						ch <- "nth"
						return nil
					}
			},
			wantVal: "tea",
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

type (
	testcaseEitherMap struct {
		title string
		arg   circle.Either
		f     func(int) (int, error)
		want  circle.Either
	}
)

func (s *testcaseEitherMap) test(t *testing.T) {
	f, err := circle.NewMapper(s.f)
	assert.Nil(t, err)
	got := s.arg.Map(f)
	{
		gotVal, gotOK := got.Right()
		wantVal, wantOK := s.want.Right()
		assert.Equal(t, gotOK, wantOK)
		assert.Equal(t, gotVal, wantVal)
	}
	{
		gotVal, gotOK := got.Left()
		wantVal, wantOK := s.want.Left()
		assert.Equal(t, gotOK, wantOK)
		assert.Equal(t, gotVal, wantVal)
	}
}

func TestEitherMap(t *testing.T) {
	for _, tc := range []*testcaseEitherMap{
		{
			title: "right ok",
			arg:   circle.NewRight(1),
			f:     func(x int) (int, error) { return x + 1, nil },
			want:  circle.NewRight(2),
		},
		{
			title: "right failure",
			arg:   circle.NewRight(1),
			f:     func(int) (int, error) { return 0, errors.New("failure") },
			want:  circle.NewLeft(errors.New("failure")),
		},
		{
			title: "left ok",
			arg:   circle.NewLeft(10),
			f:     func(int) (int, error) { return 0, nil },
			want:  circle.NewLeft(10),
		},
		{
			title: "left failure",
			arg:   circle.NewLeft(10),
			f:     func(int) (int, error) { return 0, errors.New("failure") },
			want:  circle.NewLeft(10),
		},
	} {
		t.Run(tc.title, tc.test)
	}
}

func TestTuple(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		v := circle.NewTuple()
		assert.Equal(t, 0, v.Size())
		_, ok := v.Get(0)
		assert.False(t, ok)
	})

	t.Run("double", func(t *testing.T) {
		v := circle.NewTuple(1, "two")
		assert.Equal(t, 2, v.Size())
		{
			x, ok := v.Get(0)
			assert.True(t, ok)
			assert.Equal(t, 1, x)
		}
		{
			x, ok := v.Get(1)
			assert.True(t, ok)
			assert.Equal(t, "two", x)
		}
		{
			_, ok := v.Get(2)
			assert.False(t, ok)
		}
	})
}
