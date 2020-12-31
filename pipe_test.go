package pipe

import (
	"fmt"
	"testing"
)

func intSliceEqual(a []int, b ...int) bool {
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

func strSliceEqual(a []string, b ...string) bool {
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

func TestPToSlice(t *testing.T) {
	src := []int{1, 2, 3}
	dst := NewPipe(src).
		Map(func(item int) string { return fmt.Sprintf("#%d", item) }).
		PToSlice().([]string)
	if !strSliceEqual(dst, "#1", "#2", "#3") {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestUniq(t *testing.T) {
	src := []int{1, 2, 1, 2, 3}
	dst := NewPipe(src).
		Uniq().
		ToSlice().([]int)
	if !intSliceEqual(dst, 1, 2, 3) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestMap(t *testing.T) {
	src := []int{1, 2, 3}
	dst := NewPipe(src).
		Map(func(item int) string { return fmt.Sprintf("#%d", item) }).
		ToSlice().([]string)
	if !strSliceEqual(dst, "#1", "#2", "#3") {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestMap2(t *testing.T) {
	src := []int{1, 2, 3}
	dst := NewPipe(src).
		Map(nil).
		ToSlice().([]int)
	if !intSliceEqual(dst, 1, 2, 3) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

type User struct {
	UserId int
}

func TestMap3(t *testing.T) {
	src := []User{User{UserId: 1}, User{UserId: 2}}
	dst := NewPipe(src).
		Map(func(item interface{}) int { return item.(User).UserId }).
		ToSlice().([]int)
	if !intSliceEqual(dst, 1, 2) {
		t.Error(fmt.Sprintf("wrong dst %v", dst))
	}
}

func TestFilter(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := NewPipe(src).
		Filter(func(in int) bool { return in%3 == 0 }).
		ToSlice().([]int)
	if !intSliceEqual(dst, 3, 6, 9) {
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

func TestPReduce(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := NewPipe(src).
		PReduce(0, func(s, item int) int { return s + item }).(int)
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
	if !intSliceEqual(dst, 9, 36, 81) {
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
	if !intSliceEqual(dst, 1, 1, 3, 4, 5, 9) {
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
	if !intSliceEqual(dst, 1, 9, 25) {
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
	if !intSliceEqual(dst, 2, 3, 4) {
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
		t.Error("ToMap: to map fail. len not matched", len(dst), 5)
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		v := fmt.Sprintf("Val-%d", i)
		if dst[k] != v {
			t.Error(fmt.Sprintf("ToMap: value wrong m[%s] = %s != %s", k, dst[k], v))
		}
	}
	dst = NewPipe(src).
		PToMap(
			func(v int) string { return fmt.Sprintf("Key-%d", v) },
			func(v int) string { return fmt.Sprintf("Val-%d", v) },
		).(map[string]string)
	if len(dst) != 5 {
		t.Error("PToMap: to map fail. len not matched", len(dst), 5)
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		v := fmt.Sprintf("Val-%d", i)
		if dst[k] != v {
			t.Error(fmt.Sprintf("PToMap: value wrong m[%s] = %s != %s", k, dst[k], v))
		}
	}
}

func TestToMapNil(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		ToMap(
			func(v int) string { return fmt.Sprintf("Key-%d", v) },
			nil,
		).(map[string]int)
	if len(dst) != 5 {
		t.Error("to map fail. len not matched")
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		if dst[k] != i {
			t.Error("value wrong")
		}
	}
}

func TestToMap2(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		ToMap2(func(v int) (string, string) {
			return fmt.Sprintf("Key-%d", v), fmt.Sprintf("Val-%d", v)
		}).(map[string]string)
	if len(dst) != 5 {
		t.Error("ToMap2: to map fail. len not matched", len(dst), 5)
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		v := fmt.Sprintf("Val-%d", i)
		if dst[k] != v {
			t.Error("ToMap2: value wrong", k, dst[k], v)
		}
	}
	dst = NewPipe(src).
		PToMap2(func(v int) (string, string) {
			return fmt.Sprintf("Key-%d", v), fmt.Sprintf("Val-%d", v)
		}).(map[string]string)
	if len(dst) != 5 {
		t.Error("PToMap2: to map fail. len not matched", len(dst), 5)
	}
	for i := 1; i <= 5; i++ {
		k := fmt.Sprintf("Key-%d", i)
		v := fmt.Sprintf("Val-%d", i)
		if dst[k] != v {
			t.Error("PToMap2: value wrong", k, dst[k], v)
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
	if !intSliceEqual(dst["odd"], 5, 3, 1) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], 4, 2) {
		t.Error("to groupmap fail.")
	}
	dst = NewPipe(src).
		PToGroupMap(
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
	if !intSliceEqual(dst["odd"], 5, 3, 1) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], 4, 2) {
		t.Error("to groupmap fail.")
	}
}

func TestToGroupMapNil(t *testing.T) {
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
			nil,
		).(map[string][]int)
	if len(dst) != 2 {
		t.Error("to groupmap fail. len not matched")
	}
	if !intSliceEqual(dst["odd"], 5, 3, 1) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], 4, 2) {
		t.Error("to groupmap fail.")
	}
}

func TestToGroupMap2(t *testing.T) {
	src := []int{5, 4, 3, 2, 1}
	dst := NewPipe(src).
		ToGroupMap2(
			func(v int) (string, int) {
				if v%2 != 0 {
					return "odd", v
				} else {
					return "even", v
				}
			},
		).(map[string][]int)
	if len(dst) != 2 {
		t.Error("to groupmap fail. len not matched")
	}
	if !intSliceEqual(dst["odd"], 5, 3, 1) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], 4, 2) {
		t.Error("to groupmap fail.")
	}
	dst = NewPipe(src).
		PToGroupMap2(
			func(v int) (string, int) {
				if v%2 != 0 {
					return "odd", v
				} else {
					return "even", v
				}
			},
		).(map[string][]int)
	if len(dst) != 2 {
		t.Error("to groupmap fail. len not matched")
	}
	if !intSliceEqual(dst["odd"], 5, 3, 1) {
		t.Error("to groupmap fail.")
	}
	if !intSliceEqual(dst["even"], 4, 2) {
		t.Error("to groupmap fail.")
	}
}

func TestMapPipeKeys(t *testing.T) {
	src := map[string]int{
		"Andy":  91,
		"Bob":   87,
		"Clark": 95,
	}
	dst := NewMapPipe(src).Keys().Sort(func(a, b string) bool { return a < b }).
		ToSlice().([]string)
	if !strSliceEqual(dst, "Andy", "Bob", "Clark") {
		t.Error("keys wrong")
	}
}

func TestMapPipeValues(t *testing.T) {
	src := map[string]int{
		"Andy":  91,
		"Bob":   87,
		"Clark": 95,
	}
	dst := NewMapPipe(src).Values().Sort(func(a, b int) bool { return a < b }).
		ToSlice().([]int)
	if !intSliceEqual(dst, 87, 91, 95) {
		t.Error("values wrong")
	}
}

func TestEach(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 5)
	NewPipe(src).
		Map(func(i int) int { return i * i }).
		Each(func(item, index int) { dst[index] = item })
	if !intSliceEqual(dst, 1, 4, 9, 16, 25) {
		t.Error("values wrong")
	}
}

func TestPEach(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 5)
	NewPipe(src).
		Map(func(i int) int { return i * i }).
		PEach(func(item, index int) { dst[index] = item })
	if !intSliceEqual(dst, 1, 4, 9, 16, 25) {
		t.Error("values wrong")
	}
}

func TestSome1(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	lessThen := func(max int) func(int) bool {
		return func(v int) bool { return v < max }
	}
	if NewPipe(src).Some(lessThen(5), 5) {
		t.Error("Some(lessThen(5), 5) shoud be false")
	}
	if !NewPipe(src).Some(lessThen(5), 4) {
		t.Error("Some(lessThen(5), 4) shoud be true")
	}
}

func TestEvery(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	lessThen := func(max int) func(int) bool {
		return func(v int) bool { return v < max }
	}
	if NewPipe(src).Every(lessThen(5)) {
		t.Error("Every(lessThen(5)) shoud be false")
	}
	if !NewPipe(src).Every(lessThen(11)) {
		t.Error("Every(lessThen(11)) shoud be true")
	}
}

func TestRange(t *testing.T) {
	if dst := Range1(5).ToSlice().([]int); !intSliceEqual(dst, 0, 1, 2, 3, 4) {
		t.Error("Range1 error", dst)
	}
	if dst := Range2(2, 5).ToSlice().([]int); !intSliceEqual(dst, 2, 3, 4) {
		t.Error("Range2 error", dst)
	}
	if dst := Range3(1, 6, 2).ToSlice().([]int); !intSliceEqual(dst, 1, 3, 5) {
		t.Error("Range3 error", dst)
	}
	if dst := Range3(1, 5, 2).ToSlice().([]int); !intSliceEqual(dst, 1, 3) {
		t.Error("Range3 error", dst)
	}
	if dst := Range(5).ToSlice().([]int); !intSliceEqual(dst, 0, 1, 2, 3, 4) {
		t.Error("Range for Range1 error", dst)
	}
	if dst := Range(2, 5).ToSlice().([]int); !intSliceEqual(dst, 2, 3, 4) {
		t.Error("Range for Range2 error", dst)
	}
	if dst := Range(1, 6, 2).ToSlice().([]int); !intSliceEqual(dst, 1, 3, 5) {
		t.Error("Range for Range3 error", dst)
	}
	if dst := Range(1, 5, 2).ToSlice().([]int); !intSliceEqual(dst, 1, 3) {
		t.Error("Range for Range3 error", dst)
	}
}

func TestRangeMapFilter(t *testing.T) {
	dst := Range1(10).
		Map(func(i int) int { return i * 2 }).
		Filter(func(i int) bool { return i < 10 }).
		ToSlice().([]int)
	if !intSliceEqual(dst, 0, 2, 4, 6, 8) {
		t.Error(dst)
	}
}

func longTimeProc(n int) int {
	if n <= 2 {
		return 1
	}
	return longTimeProc(n-1) + longTimeProc(n-2)
}

const N = 35

func BenchmarkEach(b *testing.B) {
	src := []int{N, N, N, N, N}
	var results [5]int
	NewPipe(src).Each(func(item, index int) {
		results[index] = longTimeProc(item)
	})
}

func BenchmarkPEach(b *testing.B) {
	src := []int{N, N, N, N, N}
	var results [5]int
	NewPipe(src).PEach(func(item, index int) {
		results[index] = longTimeProc(item)
	})
}

func BenchmarkToSlice(b *testing.B) {
	src := []int{N, N, N, N, N}
	NewPipe(src).Map(longTimeProc).ToSlice()
}

func BenchmarkPToSlice(b *testing.B) {
	src := []int{N, N, N, N, N}
	NewPipe(src).Map(longTimeProc).PToSlice()
}

func sumIntReducer(sum, input int) int {
	return sum + input
}

func BenchmarkReduce(b *testing.B) {
	src := []int{N, N, N, N, N}
	NewPipe(src).Map(longTimeProc).Reduce(0, sumIntReducer)
}

func BenchmarkPReduce(b *testing.B) {
	src := []int{N, N, N, N, N}
	NewPipe(src).Map(longTimeProc).PReduce(0, sumIntReducer)
}
