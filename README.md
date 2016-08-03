# A golang library that makes operations on slice easilier

## What can I do?
* data process
  * Map
  * Filter
  * Sort
  * Reverse
* output
  * ToSlice
  * ToMap
  * Reduce

## Example
```go
	src := []int{1, 2, 3}
	dst := pipe.NewPipe(src).
			Map(func(item int) string{ return fmt.Sprintf("#%d", item) }).
			ToSlice().([]string)
	// dst is []string{"#1", "#2", "#3"}
```
```go
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := pipe.NewPipe(src).
			Filter(func(in int) bool { return in % 3 == 0 }).
			ToSlice().([]int)
	// dst is []int{3, 6, 9}
```
```go
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := pipe.NewPipe(src).
			Reduce(0, func(s, item int) int{ return s + item }).(int)
	// sum == 55
```
```go
	src := []int{3, 1, 4, 1, 5, 9}
	dst := pipe.NewPipe(src).
		Sort(func(a, b int) bool { return a < b }).
		ToSlice().([]int)
	// dst is []int{1, 1, 3, 4, 5, 9}
```
```go
	src := []int{1, 2, 3}
	dst := pipe.NewPipe(src).
		Reverse().
		ToSlice().([]int)
	// dst is []int{3, 2, 1}
```
```go
	src := []int{5, 4, 3, 2, 1}
	dst := pipe.NewPipe(src).
		ToMap(
			func(v int) string { return fmt.Sprintf("Key-%d", v) },
			func(v int) string { return fmt.Sprintf("Val-%d", v) },
		).(map[string]string)
	// dst is map[Key-1:Val-1 Key-5:Val-5 Key-4:Val-4 Key-3:Val-3 Key-2:Val-2]
```
You can invoke map/filter function many times
```go
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	dst := pipe.NewPipe(src).
			Filter(func(in int) bool { return in % 3 == 0 }).
			Map(func(in int) int { return in * in }).
			ToSlice().([]int)
	// dst is []int{9, 36, 81}
```
