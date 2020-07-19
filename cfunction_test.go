package circle_test

import (
	"circle"
	"errors"
	"fmt"
	"testing"

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
			title:        "not tuple",
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
