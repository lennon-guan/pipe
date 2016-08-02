# A golang library that makes operations on slice easilier

## What can I do?
* map
* filter
* reduce
* sort
* reverse

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
	dst := NewPipe(src).
		Sort(func(a, b int) bool { return a < b }).
		ToSlice().([]int)
	// dst is []int{1, 1, 3, 4, 5, 9}
```
```go
	src := []int{1, 2, 3}
	dst := NewPipe(src).
		Reverse().
		ToSlice().([]int)
	// dst is []int{3, 2, 1}
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
