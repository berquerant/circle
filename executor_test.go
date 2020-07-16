package circle_test

import (
	"circle"
	"errors"
	"fmt"
	"strings"
	"testing"

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
