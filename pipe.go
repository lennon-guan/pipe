package pipe

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type _IProc interface {
	Next(reflect.Value) (reflect.Value, bool)
	GetOutType() reflect.Type
}

func isTypeMatched(a, b reflect.Type) bool {
	if a == b {
		return true
	} else if a.Kind() == reflect.Interface {
		return b.Implements(a)
	}
	return false
}

func isGoodFunc(fType reflect.Type, intypes, outtypes []interface{}) bool {
	if fType.Kind() != reflect.Func {
		return false
	}
	if fType.NumIn() != len(intypes) {
		return false
	}
	for i, t := range intypes {
		argType := fType.In(i)
		if t == nil {
			continue
		}
		if tt, ok := t.(reflect.Type); ok && !isTypeMatched(tt, argType) {
			return false
		}
		if tk, ok := t.(reflect.Kind); ok && tk != argType.Kind() {
			return false
		}
	}
	if fType.NumOut() != len(outtypes) {
		return false
	}
	for i, t := range outtypes {
		argType := fType.Out(i)
		if t == nil {
			continue
		}
		if tt, ok := t.(reflect.Type); ok && !isTypeMatched(tt, argType) {
			return false
		}
		if tk, ok := t.(reflect.Kind); ok && tk != argType.Kind() {
			return false
		}
	}
	return true
}

func noop(input reflect.Value) reflect.Value {
	return input
}

type _MapProc struct {
	InType  reflect.Type
	OutType reflect.Type
	Func    reflect.Value
}

func newMapProc(f interface{}, intype reflect.Type) *_MapProc {
	fValue := reflect.ValueOf(f)
	if !fValue.IsValid() {
		return &_MapProc{
			InType:  intype,
			OutType: intype,
			Func:    fValue,
		}
	} else {
		fType := fValue.Type()
		if !isGoodFunc(fType, []interface{}{nil}, []interface{}{nil}) {
			panic("map function must has only one input parameter and only one output parameter")
		}
		return &_MapProc{
			InType:  fType.In(0),
			OutType: fType.Out(0),
			Func:    fValue,
		}
	}
}

func (m *_MapProc) Next(input reflect.Value) (reflect.Value, bool) {
	inType := input.Type()
	if !isTypeMatched(m.InType, inType) {
		panic(fmt.Sprintf("input type error. want %v got %v", m.InType, inType))
	}
	if m.Func.IsValid() {
		outs := m.Func.Call([]reflect.Value{input})
		return outs[0], true
	} else {
		return input, true
	}
}

func (m *_MapProc) GetOutType() reflect.Type {
	return m.OutType
}

type _FilterProc struct {
	InType  reflect.Type
	OutType reflect.Type
	Func    reflect.Value
}

func newFilterProc(f interface{}) *_FilterProc {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if !isGoodFunc(fType, []interface{}{nil}, []interface{}{reflect.Bool}) {
		panic("filter function must has only one input parameter and only one boolean output parameter")
	}
	return &_FilterProc{
		InType:  fType.In(0),
		OutType: fType.In(0),
		Func:    fValue,
	}
}

func (f *_FilterProc) Next(input reflect.Value) (reflect.Value, bool) {
	inType := input.Type()
	if !isTypeMatched(f.InType, inType) {
		panic(fmt.Sprintf("input type error. want %v got %v", f.InType, inType))
	}
	outs := f.Func.Call([]reflect.Value{input})
	passed := outs[0].Interface().(bool)
	return input, passed
}

func (f *_FilterProc) GetOutType() reflect.Type {
	return f.OutType
}

type _Pipe struct {
	arr     interface{}
	srcPipe *_Pipe
	proc    _IProc
}

func NewPipe(arr interface{}) *_Pipe {
	return &_Pipe{
		arr:     arr,
		srcPipe: nil,
		proc:    nil,
	}
}

func (p *_Pipe) Filter(proc interface{}) *_Pipe {
	return &_Pipe{
		arr:     nil,
		srcPipe: p,
		proc:    newFilterProc(proc),
	}
}

func (p *_Pipe) Map(proc interface{}) *_Pipe {
	return &_Pipe{
		arr:     nil,
		srcPipe: p,
		proc:    newMapProc(proc, p.getOutType()),
	}
}

func (p *_Pipe) srcLen() int {
	pp := p
	for pp.srcPipe != nil {
		pp = pp.srcPipe
	}
	if pp.arr == nil {
		panic("no slice")
	}
	return reflect.ValueOf(pp.arr).Len()
}

type _GetValueTask struct {
	srcIndex   int
	startValue reflect.Value
	procList   []_IProc
}

func (t *_GetValueTask) GetValue() (item reflect.Value, keep bool) {
	item = t.startValue
	keep = true
	if t.procList != nil {
		for _, proc := range t.procList {
			item, keep = proc.Next(item)
			if !keep {
				return
			}
		}
	}
	return
}

func (t *_GetValueTask) PGetValue() (item reflect.Value, keep bool) {
	item = t.startValue
	keep = true
	if t.procList != nil {
		done := make(chan int, 1)
		go func() {
			for _, proc := range t.procList {
				item, keep = proc.Next(item)
				if !keep {
					break
				}
			}
			done <- 1
		}()
		<-done
		close(done)
	}
	return
}

func (t *_GetValueTask) Then(fn func(reflect.Value)) {
	item, keep := t.PGetValue()
	if keep {
		fn(item)
	}
}

func (t *_GetValueTask) UnstableThen(wg *sync.WaitGroup, fn func(reflect.Value)) {
	go func() {
		defer wg.Done()
		item, keep := t.PGetValue()
		if keep {
			fn(item)
		}
	}()
}

func (t *_GetValueTask) StableThen(wait *WaitIndex, fn func(reflect.Value)) {
	go func() {
		item, keep := t.PGetValue()
		if keep {
			wait.Wait(t.srcIndex - 1)
			fn(item)
		}
		wait.Done(t.srcIndex)
	}()
}

func (p *_Pipe) getValue(index int) (task *_GetValueTask) {
	if p.srcPipe != nil {
		task = p.srcPipe.getValue(index)
	} else {
		task = &_GetValueTask{
			srcIndex:   index,
			startValue: reflect.ValueOf(p.arr).Index(index),
		}
	}
	if p.proc != nil {
		if task.procList == nil {
			task.procList = make([]_IProc, 1, 10)
			task.procList[0] = p.proc
		} else {
			task.procList = append(task.procList, p.proc)
		}
	}
	return
}

func (p *_Pipe) getOutType() reflect.Type {
	if p.proc != nil {
		return p.proc.GetOutType()
	} else if p.arr != nil {
		return reflect.TypeOf(p.arr).Elem()
	} else {
		panic("both proc and arr are nil")
	}
}

func (p *_Pipe) ToSlice() interface{} {
	if p.proc == nil {
		return p.arr
	}
	length := p.srcLen()
	outElemType := p.getOutType()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			newSliceValue = reflect.Append(newSliceValue, itemValue)
		})
	}
	return newSliceValue.Interface()
}

func (p *_Pipe) PToSlice() interface{} {
	if p.proc == nil {
		return p.arr
	}
	length := p.srcLen()
	outElemType := p.getOutType()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	waitIndex := NewWaitIndex(length)
	for i := 0; i < length; i++ {
		p.getValue(i).StableThen(waitIndex, func(itemValue reflect.Value) {
			newSliceValue = reflect.Append(newSliceValue, itemValue)
		})
	}
	waitIndex.WaitAndClose()
	return newSliceValue.Interface()
}

func (p *_Pipe) Each(fn interface{}) {
	fType := reflect.TypeOf(fn)
	if !isGoodFunc(fType, []interface{}{p.getOutType(), reflect.Int}, []interface{}{}) {
		panic("Invalid each process function")
	}
	fValue := reflect.ValueOf(fn)
	length := p.srcLen()
	index := 0
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			fValue.Call([]reflect.Value{itemValue, reflect.ValueOf(index)})
			index++
		})
	}
}

func (p *_Pipe) PEach(fn interface{}) {
	fType := reflect.TypeOf(fn)
	if !isGoodFunc(fType, []interface{}{p.getOutType(), reflect.Int}, []interface{}{}) {
		panic("Invalid each process function")
	}
	fValue := reflect.ValueOf(fn)
	var wg sync.WaitGroup
	length := p.srcLen()
	index := 0
	wi := NewWaitIndex(length)
	for i := 0; i < length; i++ {
		p.getValue(i).StableThen(wi, func(itemValue reflect.Value) {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				fValue.Call([]reflect.Value{itemValue, reflect.ValueOf(index)})
			}(index)
			index++
		})
	}
	wi.WaitAndClose()
	wg.Wait()
}

func (p *_Pipe) ToMap(getKey, getVal interface{}) interface{} {
	outElemType := p.getOutType()
	var keyType, valType reflect.Type
	getKeyValue := reflect.ValueOf(getKey)
	var realGetKey, realGetVal func(reflect.Value) reflect.Value
	if getKeyValue.IsValid() {
		getKeyType := getKeyValue.Type()
		if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getKey func invalid")
		}
		keyType = getKeyType.Out(0)
		realGetKey = func(input reflect.Value) reflect.Value {
			return getKeyValue.Call([]reflect.Value{input})[0]
		}
	} else {
		keyType = outElemType
		realGetVal = noop
	}
	getValValue := reflect.ValueOf(getVal)
	if getValValue.IsValid() {
		getValType := getValValue.Type()
		if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getVal func invalid")
		}
		valType = getValType.Out(0)
		realGetVal = func(input reflect.Value) reflect.Value {
			return getValValue.Call([]reflect.Value{input})[0]
		}
	} else {
		valType = outElemType
		realGetVal = noop
	}
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, valType))
	length := p.srcLen()
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			newMapValue.SetMapIndex(realGetKey(itemValue), realGetVal(itemValue))
		})
	}
	return newMapValue.Interface()
}

func (p *_Pipe) PToMap(getKey, getVal interface{}) interface{} {
	outElemType := p.getOutType()
	var keyType, valType reflect.Type
	getKeyValue := reflect.ValueOf(getKey)
	var realGetKey, realGetVal func(reflect.Value) reflect.Value
	if getKeyValue.IsValid() {
		getKeyType := getKeyValue.Type()
		if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getKey func invalid")
		}
		keyType = getKeyType.Out(0)
		realGetKey = func(input reflect.Value) reflect.Value {
			return getKeyValue.Call([]reflect.Value{input})[0]
		}
	} else {
		keyType = outElemType
		realGetVal = noop
	}
	getValValue := reflect.ValueOf(getVal)
	if getValValue.IsValid() {
		getValType := getValValue.Type()
		if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getVal func invalid")
		}
		valType = getValType.Out(0)
		realGetVal = func(input reflect.Value) reflect.Value {
			return getValValue.Call([]reflect.Value{input})[0]
		}
	} else {
		valType = outElemType
		realGetVal = noop
	}
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, valType))
	length := p.srcLen()
	var wg sync.WaitGroup
	wg.Add(length)
	var lock sync.Mutex
	for i := 0; i < length; i++ {
		p.getValue(i).UnstableThen(&wg, func(itemValue reflect.Value) {
			lock.Lock()
			newMapValue.SetMapIndex(realGetKey(itemValue), realGetVal(itemValue))
			lock.Unlock()
		})
	}
	wg.Wait()
	return newMapValue.Interface()
}

func (p *_Pipe) ToMap2(getPair interface{}) interface{} {
	getPairValue := reflect.ValueOf(getPair)
	getPairType := getPairValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getPairType, []interface{}{outElemType}, []interface{}{nil, nil}) {
		panic("getPair func invalid")
	}
	keyType := getPairType.Out(0)
	valType := getPairType.Out(1)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, valType))
	length := p.srcLen()
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			outs := getPairValue.Call([]reflect.Value{itemValue})
			newMapValue.SetMapIndex(outs[0], outs[1])
		})
	}
	return newMapValue.Interface()
}

func (p *_Pipe) PToMap2(getPair interface{}) interface{} {
	getPairValue := reflect.ValueOf(getPair)
	getPairType := getPairValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getPairType, []interface{}{outElemType}, []interface{}{nil, nil}) {
		panic("getPair func invalid")
	}
	keyType := getPairType.Out(0)
	valType := getPairType.Out(1)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, valType))
	length := p.srcLen()
	var wg sync.WaitGroup
	wg.Add(length)
	var lock sync.Mutex
	for i := 0; i < length; i++ {
		p.getValue(i).UnstableThen(&wg, func(itemValue reflect.Value) {
			outs := getPairValue.Call([]reflect.Value{itemValue})
			lock.Lock()
			newMapValue.SetMapIndex(outs[0], outs[1])
			lock.Unlock()
		})
	}
	wg.Wait()
	return newMapValue.Interface()
}

func (p *_Pipe) ToGroupMap(getKey, getVal interface{}) interface{} {
	outElemType := p.getOutType()
	var keyType, valType reflect.Type
	getKeyValue := reflect.ValueOf(getKey)
	var realGetKey, realGetVal func(reflect.Value) reflect.Value
	if getKeyValue.IsValid() {
		getKeyType := getKeyValue.Type()
		if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getKey func invalid")
		}
		keyType = getKeyType.Out(0)
		realGetKey = func(input reflect.Value) reflect.Value {
			return getKeyValue.Call([]reflect.Value{input})[0]
		}
	} else {
		keyType = outElemType
		realGetVal = noop
	}
	getValValue := reflect.ValueOf(getVal)
	if getValValue.IsValid() {
		getValType := getValValue.Type()
		if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getVal func invalid")
		}
		valType = getValType.Out(0)
		realGetVal = func(input reflect.Value) reflect.Value {
			return getValValue.Call([]reflect.Value{input})[0]
		}
	} else {
		valType = outElemType
		realGetVal = noop
	}
	sliceType := reflect.SliceOf(valType)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, sliceType))
	length := p.srcLen()
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			keyValue := realGetKey(itemValue)
			valValue := realGetVal(itemValue)
			slot := newMapValue.MapIndex(keyValue)
			if !slot.IsValid() {
				slot = reflect.MakeSlice(sliceType, 0, length-i)
			}
			slot = reflect.Append(slot, valValue)
			newMapValue.SetMapIndex(keyValue, slot)
		})
	}
	return newMapValue.Interface()
}

func (p *_Pipe) PToGroupMap(getKey, getVal interface{}) interface{} {
	outElemType := p.getOutType()
	var keyType, valType reflect.Type
	getKeyValue := reflect.ValueOf(getKey)
	var realGetKey, realGetVal func(reflect.Value) reflect.Value
	if getKeyValue.IsValid() {
		getKeyType := getKeyValue.Type()
		if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getKey func invalid")
		}
		keyType = getKeyType.Out(0)
		realGetKey = func(input reflect.Value) reflect.Value {
			return getKeyValue.Call([]reflect.Value{input})[0]
		}
	} else {
		keyType = outElemType
		realGetVal = noop
	}
	getValValue := reflect.ValueOf(getVal)
	if getValValue.IsValid() {
		getValType := getValValue.Type()
		if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
			panic("getVal func invalid")
		}
		valType = getValType.Out(0)
		realGetVal = func(input reflect.Value) reflect.Value {
			return getValValue.Call([]reflect.Value{input})[0]
		}
	} else {
		valType = outElemType
		realGetVal = noop
	}
	sliceType := reflect.SliceOf(valType)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, sliceType))
	length := p.srcLen()
	wi := NewWaitIndex(length)
	for i := 0; i < length; i++ {
		p.getValue(i).StableThen(wi, func(itemValue reflect.Value) {
			keyValue := realGetKey(itemValue)
			valValue := realGetVal(itemValue)
			slot := newMapValue.MapIndex(keyValue)
			if !slot.IsValid() {
				slot = reflect.MakeSlice(sliceType, 0, length-i)
			}
			slot = reflect.Append(slot, valValue)
			newMapValue.SetMapIndex(keyValue, slot)
		})
	}
	wi.WaitAndClose()
	return newMapValue.Interface()
}

func (p *_Pipe) ToGroupMap2(getPair interface{}) interface{} {
	getPairValue := reflect.ValueOf(getPair)
	getPairType := getPairValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getPairType, []interface{}{outElemType}, []interface{}{nil, nil}) {
		panic("getPair func invalid")
	}
	keyType := getPairType.Out(0)
	valType := getPairType.Out(1)
	sliceType := reflect.SliceOf(valType)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, sliceType))
	length := p.srcLen()
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			outs := getPairValue.Call([]reflect.Value{itemValue})
			keyValue := outs[0]
			valValue := outs[1]
			slot := newMapValue.MapIndex(keyValue)
			if !slot.IsValid() {
				slot = reflect.MakeSlice(sliceType, 0, length-i)
			}
			slot = reflect.Append(slot, valValue)
			newMapValue.SetMapIndex(keyValue, slot)
		})
	}
	return newMapValue.Interface()
}

func (p *_Pipe) PToGroupMap2(getPair interface{}) interface{} {
	getPairValue := reflect.ValueOf(getPair)
	getPairType := getPairValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getPairType, []interface{}{outElemType}, []interface{}{nil, nil}) {
		panic("getPair func invalid")
	}
	keyType := getPairType.Out(0)
	valType := getPairType.Out(1)
	sliceType := reflect.SliceOf(valType)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, sliceType))
	length := p.srcLen()
	wi := NewWaitIndex(length)
	for i := 0; i < length; i++ {
		p.getValue(i).StableThen(wi, func(itemValue reflect.Value) {
			outs := getPairValue.Call([]reflect.Value{itemValue})
			keyValue := outs[0]
			valValue := outs[1]
			slot := newMapValue.MapIndex(keyValue)
			if !slot.IsValid() {
				slot = reflect.MakeSlice(sliceType, 0, length-i)
			}
			slot = reflect.Append(slot, valValue)
			newMapValue.SetMapIndex(keyValue, slot)
		})
	}
	wi.WaitAndClose()
	return newMapValue.Interface()
}

func (p *_Pipe) Reduce(initValue interface{}, proc interface{}) interface{} {
	length := p.srcLen()
	outElemType := p.getOutType()
	procValue := reflect.ValueOf(proc)
	procType := procValue.Type()
	initType := reflect.TypeOf(initValue)
	if !isGoodFunc(procType, []interface{}{initType, outElemType}, []interface{}{initType}) {
		panic("reduce function invalid")
	}
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			outs := procValue.Call([]reflect.Value{reflect.ValueOf(initValue), itemValue})
			initValue = outs[0].Interface()
		})
	}
	return initValue
}

func (p *_Pipe) PReduce(initValue interface{}, proc interface{}) interface{} {
	length := p.srcLen()
	outElemType := p.getOutType()
	procValue := reflect.ValueOf(proc)
	procType := procValue.Type()
	initType := reflect.TypeOf(initValue)
	if !isGoodFunc(procType, []interface{}{initType, outElemType}, []interface{}{initType}) {
		panic("reduce function invalid")
	}
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(length)
	for i := 0; i < length; i++ {
		p.getValue(i).UnstableThen(&wg, func(itemValue reflect.Value) {
			lock.Lock()
			defer lock.Unlock()
			outs := procValue.Call([]reflect.Value{reflect.ValueOf(initValue), itemValue})
			initValue = outs[0].Interface()
		})
	}
	wg.Wait()
	return initValue
}

type _SortDelegate struct {
	Arr      reflect.Value
	lessFunc reflect.Value
}

func (s *_SortDelegate) Len() int {
	return s.Arr.Len()
}

func (s *_SortDelegate) Less(i, j int) bool {
	outs := s.lessFunc.Call([]reflect.Value{s.Arr.Index(i), s.Arr.Index(j)})
	return outs[0].Interface().(bool)
}

func (s *_SortDelegate) Swap(i, j int) {
	ti := s.Arr.Index(i).Interface()
	tj := s.Arr.Index(j).Interface()
	s.Arr.Index(i).Set(reflect.ValueOf(tj))
	s.Arr.Index(j).Set(reflect.ValueOf(ti))
}

func (p *_Pipe) Sort(less interface{}) *_Pipe {
	lessValue := reflect.ValueOf(less)
	lessType := lessValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(lessType, []interface{}{outElemType, outElemType}, []interface{}{reflect.Bool}) {
		panic("sort less function invalid")
	}
	delegate := &_SortDelegate{
		Arr:      reflect.ValueOf(p.ToSlice()),
		lessFunc: lessValue,
	}
	sort.Stable(delegate)
	return &_Pipe{
		arr: delegate.Arr.Interface(),
	}
}

func (p *_Pipe) Uniq() *_Pipe {
	outElemType := p.getOutType()
	length := p.srcLen()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	existsValues := make(map[interface{}]int)
	for i := 0; i < length; i++ {
		p.getValue(i).Then(func(itemValue reflect.Value) {
			val := itemValue.Interface()
			if _, exists := existsValues[val]; !exists {
				existsValues[val] = 1
				newSliceValue = reflect.Append(newSliceValue, itemValue)
			}
		})
	}
	return &_Pipe{
		arr: newSliceValue.Interface(),
	}
}

func (p *_Pipe) Reverse() *_Pipe {
	outElemType := p.getOutType()
	length := p.srcLen()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	for i := length - 1; i >= 0; i-- {
		itemValue, keep := p.getValue(i).GetValue()
		if keep {
			newSliceValue = reflect.Append(newSliceValue, itemValue)
		}
	}
	return &_Pipe{
		arr: newSliceValue.Interface(),
	}
}
