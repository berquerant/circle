package circle_test

import (
	"circle"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func ExampleNewMapExecutor() {
	it, _ := circle.NewIterator([]string{"a", "ring", "bug", "a", "ring", "bug", "of", "roses"})
	f, _ := circle.NewMapper(func(x string) (string, error) {
		if x == "bug" {
			return "", errors.New("bug")
		}
		return strings.ToUpper(x), nil
	})
	exit, _ := circle.NewMapExecutor(f, it).Execute()
	c := exit.Channel()
	for v := range c.C() {
		fmt.Println(v)
	}
	if err := c.Err(); err != nil {
		panic(err)
	}
	// Output:
	// A
	// RING
	// A
	// RING
	// OF
	// ROSES
}

func TestMapExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{-1, 0, 1})
	assert.Nil(t, err)
	f, err := circle.NewMapper(func(x int) (int, error) {
		if x < 0 {
			return 0, errors.New("negative")
		}
		return x + 10, nil
	})
	assert.Nil(t, err)
	exit, err := circle.NewMapExecutor(f, it).Execute()
	assert.Nil(t, err)
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 10, v)
	}
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 11, v)
	}
	{
		_, err := exit.Next()
		assert.Equal(t, circle.ErrEOI, err)
	}
}

func ExampleNewFilterExecutor() {
	it, _ := circle.NewIterator([]string{"a", "ring", "bug", "a", "ring", "stop", "of", "roses"})
	f, _ := circle.NewFilter(func(x string) (bool, error) {
		if x == "stop" {
			return false, errors.New("stop")
		}
		return x != "bug", nil
	})
	exit, _ := circle.NewFilterExecutor(f, it).Execute()
	c := exit.Channel()
	for v := range c.C() {
		fmt.Println(v)
	}
	fmt.Println(c.Err())
	// Output:
	// a
	// ring
	// a
	// ring
	// stop
}

func TestFilterExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{1, 2, 3, -1, 9})
	assert.Nil(t, err)
	f, err := circle.NewFilter(func(x int) (bool, error) {
		if x < 0 {
			return false, errors.New("negative")
		}
		return x&1 == 1, nil
	})
	assert.Nil(t, err)
	exit, err := circle.NewFilterExecutor(f, it).Execute()
	assert.Nil(t, err)
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 1, v)
	}
	{
		v, err := exit.Next()
		assert.Nil(t, err)
		assert.Equal(t, 3, v)
	}
	{
		_, err := exit.Next()
		assert.Equal(t, errors.New("negative"), err)
	}
}

func ExampleNewAggregateExecutor_right() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	f, _ := circle.NewAggregator(func(x int, y string) (string, error) {
		return fmt.Sprintf("(%d+%s)", x, y), nil
	})
	ex, _ := circle.NewAggregateExecutor(f, it, "iv")
	exit, _ := ex.Execute()
	v, _ := exit.Next()
	fmt.Println(v)
	_, err := exit.Next()
	fmt.Println(err)
	// Output:
	// (1+(2+(3+iv)))
	// EOI
}

func ExampleNewAggregateExecutor_left() {
	it, _ := circle.NewIterator([]int{1, 2, 3})
	f, _ := circle.NewAggregator(func(x string, y int) (string, error) {
		return fmt.Sprintf("(%s+%d)", x, y), nil
	})
	ex, _ := circle.NewAggregateExecutor(f, it, "iv")
	exit, _ := ex.Execute()
	v, _ := exit.Next()
	fmt.Println(v)
	_, err := exit.Next()
	fmt.Println(err)
	// Output:
	// (((iv+1)+2)+3)
	// EOI
}

func TestAggregateExecutor(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"right": testRightAggregateExecutor,
		"left":  testLeftAggregateExecutor,
	} {
		t.Run(name, tc)
	}
}

func testLeftAggregateExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{10, 9, 4})
	assert.Nil(t, err)
	f, err := circle.NewAggregator(func(x string, y int) (string, error) {
		return fmt.Sprintf("%s+%d", x, y), nil
	})
	assert.Nil(t, err)
	ex, err := circle.NewAggregateExecutor(f, it, "0")
	assert.Nil(t, err)
	exit, err := ex.Execute()
	assert.Nil(t, err)
	v, err := exit.Next()
	assert.Nil(t, err)
	assert.Equal(t, "0+10+9+4", v)
	_, err = exit.Next()
	assert.Equal(t, circle.ErrEOI, err)
}

func testRightAggregateExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{10, 9, 4})
	assert.Nil(t, err)
	f, err := circle.NewAggregator(func(x, y int) (int, error) {
		return x - y, nil
	})
	assert.Nil(t, err)
	ex, err := circle.NewAggregateExecutor(f, it, 0, circle.WithAggregateExecutorType(circle.RAggregateExecutorType))
	assert.Nil(t, err)
	exit, err := ex.Execute()
	assert.Nil(t, err)
	v, err := exit.Next()
	assert.Nil(t, err)
	assert.Equal(t, 5, v)
	_, err = exit.Next()
	assert.Equal(t, circle.ErrEOI, err)
}

func TestCompareExecutor(t *testing.T) {
	it, err := circle.NewIterator([]int{5, 2, 3, 1, 4})
	assert.Nil(t, err)
	f, err := circle.NewComparator(func(x, y int) (bool, error) {
		return x < y, nil
	})
	assert.Nil(t, err)
	exit, err := circle.NewCompareExecutor(f, it).Execute()
	assert.Nil(t, err)
	c := exit.Channel()
	xs := []int{}
	for v := range c.C() {
		xs = append(xs, v.(int))
	}
	assert.Equal(t, "", cmp.Diff([]int{1, 2, 3, 4, 5}, xs))
	assert.Nil(t, c.Err())
}

func ExampleNewFlatExecutor() {
	it, _ := circle.NewIterator([][]string{[]string{"cast"}, []string{"a", "spell"}, []string{"on", "me"}})
	exit, _ := circle.NewFlatExecutor(it).Execute()
	for v := range exit.Channel().C() {
		fmt.Println(v)
	}
	// Output:
	// cast
	// a
	// spell
	// on
	// me
}

func TestFlatExecutor(t *testing.T) {
	for name, tc := range map[string]func(t *testing.T){
		"noop":    testFlatExecutorNoop,
		"onestep": testFlatExecutorFlatOneStep,
	} {
		t.Run(name, tc)
	}
}

func testFlatExecutorNoop(t *testing.T) {
	it, err := circle.NewIterator([]int{1, 2, 3})
	assert.Nil(t, err)
	exit, err := circle.NewFlatExecutor(it).Execute()
	assert.Nil(t, err)
	c := exit.Channel()
	xs := []int{}
	for v := range c.C() {
		xs = append(xs, v.(int))
	}
	assert.Equal(t, "", cmp.Diff([]int{1, 2, 3}, xs))
	assert.Nil(t, c.Err())
}

func testFlatExecutorFlatOneStep(t *testing.T) {
	it, err := circle.NewIterator([][]int{[]int{1}, []int{2, 3}, []int{4, 5, 6}})
	assert.Nil(t, err)
	exit, err := circle.NewFlatExecutor(it).Execute()
	assert.Nil(t, err)
	c := exit.Channel()
	xs := []int{}
	for v := range c.C() {
		xs = append(xs, v.(int))
	}
	assert.Equal(t, "", cmp.Diff([]int{1, 2, 3, 4, 5, 6}, xs))
	assert.Nil(t, c.Err())
}
