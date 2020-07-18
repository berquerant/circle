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
