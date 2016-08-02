package pipe

import (
	"fmt"
	"testing"
)

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		if b[i] != va {
			return false
		}
	}
	return true
}

func strSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		if b[i] != va {
			return false
		}
	}
	return true
}

func TestMap(t *testing.T) {
	src := []int{1, 2, 3}
	dst := NewPipe(src).
			Map(func(item int) string{ return fmt.Sprintf("#%d", item) }).
			ToSlice().([]string)
	if !strSliceEqual(dst, []string{"#1", "#2", "#3"}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestFilter(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := NewPipe(src).
			Filter(func(in int) bool { return in % 3 == 0 }).
			ToSlice().([]int)
	if !intSliceEqual(dst, []int{3, 6, 9}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestReduce(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := NewPipe(src).
			Reduce(0, func(s, item int) int{ return s + item }).(int)
	if sum != 55 {
		t.Error(fmt.Sprintf("sum %v != 55", sum))
	}
}

func TestMapFilterToSlice(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := NewPipe(src).
			Filter(func(in int) bool { return in % 3 == 0 }).
			Map(func(in int) int { return in * in }).
			ToSlice().([]int)
	if !intSliceEqual(dst, []int{9, 36, 81}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestMapFilterReduce(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := NewPipe(src).
			Filter(func(in int) bool { return in % 3 == 0 }).
			Map(func(in int) int { return in * in }).
			Reduce(0, func(s, item int)int { return s + item }).(int)
	if sum != 126 {
		t.Error(fmt.Sprintf("sum %v != 126", sum))
	}
}

