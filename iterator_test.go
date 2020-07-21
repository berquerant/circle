package circle_test

import (
	"circle"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func ExampleIterator() {
	it, err := circle.NewIterator([]string{"i", "t", "e", "r"})
	if err != nil {
		panic(err)
	}
	for {
		v, err := it.Next()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(v)
	}
	// Output:
	// i
	// t
	// e
	// r
	// EOI
}

func ExampleIteratorChannel() {
	it, err := circle.NewIterator([]int{1, 2, 3})
	if err != nil {
		panic(err)
	}
	c := it.Channel()
	for v := range c.C() {
		fmt.Println(v)
	}
	if err := c.Err(); err != nil {
		panic(err)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleIteratorChannel_failure() {
	var isDone bool
	it, err := circle.NewIterator(func() (interface{}, error) {
		if isDone {
			return nil, errors.New("done")
		}
		isDone = true
		return "do", nil
	})
	if err != nil {
		panic(err)
	}
	c := it.Channel()
	for v := range c.C() {
		fmt.Println(v)
	}
	fmt.Println(c.Err())
	// Output:
	// do
	// done
}

func TestIteratorChannel(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"normal":  testIteratorChannel,
		"failure": testIteratorChannelFailure,
		"context": testIteratorChannelWithContext,
	} {
		t.Run(name, tc)
	}
}

func testIteratorChannelWithContext(t *testing.T) {
	var i int
	it, err := circle.NewIterator(func() (interface{}, error) {
		// infinite iterator
		time.Sleep(100 * time.Millisecond)
		defer func() { i++ }()
		return i, nil
	})
	assert.Nil(t, err)
	var (
		ctx, cancel = context.WithTimeout(context.TODO(), 150*time.Millisecond)
		c           = it.ChannelWithContext(ctx)
		xs          = []int{}
	)
	defer cancel()
	for v := range c.C() {
		xs = append(xs, v.(int))
	}
	assert.Equal(t, "", cmp.Diff([]int{0, 1}, xs))
	assert.Nil(t, c.Err())
}

func testIteratorChannel(t *testing.T) {
	v := []int{0, 1, 2}
	it, err := circle.NewIterator(v)
	assert.Nil(t, err)
	c := it.Channel()
	got := []int{}
	for x := range c.C() {
		got = append(got, x.(int))
	}
	assert.Equal(t, "", cmp.Diff(v, got))
	assert.Nil(t, c.Err())
}

func testIteratorChannelFailure(t *testing.T) {
	e := errors.New("unknown")
	it, err := circle.NewIterator(func() (interface{}, error) {
		return nil, e
	})
	assert.Nil(t, err)
	c := it.Channel()
	var isRecv bool
	for range c.C() {
		isRecv = true
	}
	assert.False(t, isRecv)
	assert.Equal(t, e, c.Err())
}

func TestNewIterator(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"orphan":   testOrphanIterator,
		"slice":    testSliceIterator,
		"array":    testArrayIterator,
		"channel":  testChannelIterator,
		"func":     testIteratorFunc,
		"nil":      testNilIterator,
		"iterator": testIteratorFromIterator,
		"map":      testMapIterator,
	} {
		t.Run(name, tc)
	}
}

func testNilIterator(t *testing.T) {
	it, err := circle.NewIterator(nil)
	assert.Nil(t, err)
	_, err = it.Next()
	assert.Equal(t, circle.ErrEOI, err, "should yield nothing.")
}

func iteratorToInts(it circle.Iterator) ([]int, error) {
	v := []int{}
	for {
		x, err := it.Next()
		if err != nil {
			return v, err
		}
		v = append(v, x.(int))
	}
}

func testOrphanIterator(t *testing.T) {
	it, err := circle.NewIterator(1)
	assert.Nil(t, err)
	v, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 1, v)
	_, err = it.Next()
	assert.Equal(t, circle.ErrEOI, err)
}

func testIteratorFromIterator(t *testing.T) {
	v := []int{1, 2, 3}
	it, err := circle.NewIterator(v)
	assert.Nil(t, err)
	nit, err := circle.NewIterator(it)
	assert.Nil(t, err)
	got, err := iteratorToInts(nit)
	assert.Equal(t, circle.ErrEOI, err)
	assert.Equal(t, "", cmp.Diff(v, got))
}

func testSliceIterator(t *testing.T) {
	v := []int{1, 2, 3}
	it, err := circle.NewIterator(v)
	assert.Nil(t, err)
	got, err := iteratorToInts(it)
	assert.Equal(t, circle.ErrEOI, err)
	assert.Equal(t, "", cmp.Diff(v, got))
}

func testArrayIterator(t *testing.T) {
	v := [3]int{1, 2, 3}
	it, err := circle.NewIterator(v)
	assert.Nil(t, err)
	got, err := iteratorToInts(it)
	assert.Equal(t, circle.ErrEOI, err)
	assert.Equal(t, "", cmp.Diff([]int{1, 2, 3}, got))
}

func testChannelIterator(t *testing.T) {
	c := make(chan int, 3)
	for i := 0; i < 3; i++ {
		c <- i
	}
	close(c)
	it, err := circle.NewIterator(c)
	assert.Nil(t, err)
	got, err := iteratorToInts(it)
	assert.Equal(t, circle.ErrEOI, err)
	assert.Equal(t, "", cmp.Diff([]int{0, 1, 2}, got))
}

func testIteratorFunc(t *testing.T) {
	e := errors.New("error")
	var i int
	it, err := circle.NewIterator(func() (interface{}, error) {
		if i >= 3 {
			return nil, e
		}
		defer func() { i++ }()
		return i, nil
	})
	assert.Nil(t, err)
	got, err := iteratorToInts(it)
	assert.Equal(t, e, err)
	assert.Equal(t, "", cmp.Diff([]int{0, 1, 2}, got))
}

func testMapIterator(t *testing.T) {
	v := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	it, err := circle.NewIterator(v)
	assert.Nil(t, err)
	c := it.Channel()
	d := map[string]int{}
	for x := range c.C() {
		p, ok := x.(circle.Tuple)
		assert.True(t, ok)
		a, ok := p.Get(0)
		assert.True(t, ok)
		b, ok := p.Get(1)
		assert.True(t, ok)
		d[a.(string)] = b.(int)
	}
	assert.Equal(t, "", cmp.Diff(v, d))
	assert.Nil(t, c.Err())
}
