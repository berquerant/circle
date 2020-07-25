package circle_test

import (
	"errors"
	"testing"

	"github.com/berquerant/circle"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

type (
	testcaseConsumeExecutor struct {
		title       string
		src         interface{}
		f           func(chan<- interface{}) interface{}
		wantConsume []interface{}
		isError     bool
	}
)

func (s *testcaseConsumeExecutor) test(t *testing.T) {
	it, err := circle.NewIterator(s.src)
	assert.Nil(t, err)
	ch := make(chan interface{})
	cf, err := circle.NewConsumer(s.f(ch))
	assert.Nil(t, err)
	ce := circle.NewConsumeExecutor(cf, it)
	var (
		gotErr     error
		gotConsume = []interface{}{}
	)
	go func() {
		gotErr = ce.ConsumeExecute()
		close(ch)
	}()
	for x := range ch {
		gotConsume = append(gotConsume, x)
	}
	assert.Equal(t, s.isError, gotErr != nil)
	assert.Equal(t, "", cmp.Diff(s.wantConsume, gotConsume))
}

func TestConsumeExecutor(t *testing.T) {
	for _, tc := range []*testcaseConsumeExecutor{
		{
			title: "consumed",
			src:   []int{1, 2, 3},
			f: func(ch chan<- interface{}) interface{} {
				return func(x int) error {
					ch <- x
					return nil
				}
			},
			wantConsume: []interface{}{1, 2, 3},
		},
		{
			title: "stop consuming",
			src: func() circle.IteratorFunc {
				var isTail bool
				return func() (interface{}, error) {
					if isTail {
						return nil, errors.New("tail")
					}
					isTail = true
					return "1", nil
				}
			}(),
			f: func(ch chan<- interface{}) interface{} {
				return func(x string) error {
					ch <- x
					return nil
				}
			},
			wantConsume: []interface{}{"1"},
			isError:     true,
		},
		{
			title: "apply error",
			src:   []int{1, 2, 3},
			f: func(ch chan<- interface{}) interface{} {
				return func(x string) error {
					ch <- 1
					return nil
				}
			},
			wantConsume: []interface{}{},
			isError:     true,
		},
	} {
		t.Run(tc.title, tc.test)
	}
}
