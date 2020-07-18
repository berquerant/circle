package circle_test

import (
	"circle"
	"errors"
	"testing"

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
