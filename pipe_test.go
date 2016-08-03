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
		Map(func(item int) string { return fmt.Sprintf("#%d", item) }).
		ToSlice().([]string)
	if !strSliceEqual(dst, []string{"#1", "#2", "#3"}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestFilter(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := NewPipe(src).
		Filter(func(in int) bool { return in%3 == 0 }).
		ToSlice().([]int)
	if !intSliceEqual(dst, []int{3, 6, 9}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestReduce(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := NewPipe(src).
		Reduce(0, func(s, item int) int { return s + item }).(int)
	if sum != 55 {
		t.Error(fmt.Sprintf("sum %v != 55", sum))
	}
}

func TestMapFilterToSlice(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := NewPipe(src).
		Filter(func(in int) bool { return in%3 == 0 }).
		Map(func(in int) int { return in * in }).
		ToSlice().([]int)
	if !intSliceEqual(dst, []int{9, 36, 81}) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestMapFilterReduce(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := NewPipe(src).
		Filter(func(in int) bool { return in%3 == 0 }).
		Map(func(in int) int { return in * in }).
		Reduce(0, func(s, item int) int { return s + item }).(int)
	if sum != 126 {
		t.Error(fmt.Sprintf("sum %v != 126", sum))
	}
}

func TestSort(t *testing.T) {
	src := []int{3, 1, 4, 1, 5, 9}
	dst := NewPipe(src).
		Sort(func(a, b int) bool { return a < b }).
		ToSlice().([]int)
	if !intSliceEqual(dst, []int{1, 1, 3, 4, 5, 9}) {
		t.Error(fmt.Sprintf("sort fail %v", dst))
	}
}

func TestSortMapFilter(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		Filter(func(v int) bool { return v%2 != 0 }).
		Sort(func(a, b int) bool { return a < b }).
		Map(func(i int) int { return i * i }).
		ToSlice().([]int)
	if !intSliceEqual(dst, []int{1, 9, 25}) {
		t.Error(fmt.Sprintf("sort fail %v", dst))
	}
}

func TestReverse(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		Filter(func(v int) bool { return v > 1 }).
		Filter(func(v int) bool { return v < 5 }).
		Reverse().
		ToSlice().([]int)
	if !intSliceEqual(dst, []int{2, 3, 4}) {
		t.Error(fmt.Sprintf("sort fail %v", dst))
	}
}

func TestToMap(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		ToMap(
		func(v int) string { return fmt.Sprintf("Key-%d", v) },
		func(v int) string { return fmt.Sprintf("Val-%d", v) },
	).(map[string]string)
	if len(dst) != 5 {
		t.Error("to map fail. len not matched")
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		v := fmt.Sprintf("Val-%d", i)
		if dst[k] != v {
			t.Error("value wrong")
		}
	}
}

func TestToGroupMap(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		ToGroupMap(
		func(v int) string {
			if v%2 != 0 {
				return "odd"
			} else {
				return "even"
			}
		},
		func(v int) int { return v },
	).(map[string][]int)
	if len(dst) != 2 {
		t.Error("to groupmap fail. len not matched")
	}
	if !intSliceEqual(dst["odd"], []int{5, 3, 1}) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], []int{4, 2}) {
		t.Error("to groupmap fail.")
	}
}
