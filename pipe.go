package pipe

import (
	"fmt"
	"reflect"
	"sort"
)

type IProc interface {
	Next(reflect.Value) (reflect.Value, bool)
	GetOutType() reflect.Type
}

type MapProc struct {
	InType  reflect.Type
	OutType reflect.Type
	Func    reflect.Value
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
		if tt, ok := t.(reflect.Type); ok && tt != argType {
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
		if tt, ok := t.(reflect.Type); ok && tt != argType {
			return false
		}
		if tk, ok := t.(reflect.Kind); ok && tk != argType.Kind() {
			return false
		}
	}
	return true
}

func NewMapProc(f interface{}) *MapProc {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if !isGoodFunc(fType, []interface{}{nil}, []interface{}{nil}) {
		panic("map function must has only one input parameter and only one output parameter")
	}
	return &MapProc{
		InType:  fType.In(0),
		OutType: fType.Out(0),
		Func:    fValue,
	}
}

func (m *MapProc) Next(input reflect.Value) (reflect.Value, bool) {
	inType := input.Type()
	if inType != m.InType {
		panic(fmt.Sprintf("input type error. want %v got %v", m.InType, inType))
	}
	outs := m.Func.Call([]reflect.Value{input})
	return outs[0], true
}

func (m *MapProc) GetOutType() reflect.Type {
	return m.OutType
}

type FilterProc struct {
	InType  reflect.Type
	OutType reflect.Type
	Func    reflect.Value
}

func NewFilterProc(f interface{}) *FilterProc {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if !isGoodFunc(fType, []interface{}{nil}, []interface{}{reflect.Bool}) {
		panic("filter function must has only one input parameter and only one boolean output parameter")
	}
	return &FilterProc{
		InType:  fType.In(0),
		OutType: fType.In(0),
		Func:    fValue,
	}
}

func (f *FilterProc) Next(input reflect.Value) (reflect.Value, bool) {
	inType := input.Type()
	if inType != f.InType {
		panic(fmt.Sprintf("input type error. want %v got %v", f.InType, inType))
	}
	outs := f.Func.Call([]reflect.Value{input})
	passed := outs[0].Interface().(bool)
	return input, passed
}

func (f *FilterProc) GetOutType() reflect.Type {
	return f.OutType
}

type Pipe struct {
	arr     interface{}
	srcPipe *Pipe
	proc    IProc
}

func NewPipe(arr interface{}) *Pipe {
	return &Pipe{
		arr:     arr,
		srcPipe: nil,
		proc:    nil,
	}
}

func (p *Pipe) Filter(proc interface{}) *Pipe {
	return &Pipe{
		arr:     nil,
		srcPipe: p,
		proc:    NewFilterProc(proc),
	}
}

func (p *Pipe) Map(proc interface{}) *Pipe {
	return &Pipe{
		arr:     nil,
		srcPipe: p,
		proc:    NewMapProc(proc),
	}
}

func (p *Pipe) srcLen() int {
	pp := p
	for pp.srcPipe != nil {
		pp = pp.srcPipe
	}
	if pp.arr == nil {
		panic("no slice")
	}
	return reflect.ValueOf(pp.arr).Len()
}

func (p *Pipe) getValue(index int) (item reflect.Value, keep bool) {
	if p.srcPipe != nil {
		item, keep = p.srcPipe.getValue(index)
	} else {
		item = reflect.ValueOf(p.arr).Index(index)
		keep = true
	}
	if !keep {
		return
	}
	if p.proc != nil {
		item, keep = p.proc.Next(item)
	}
	return
}

func (p *Pipe) getOutType() reflect.Type {
	if p.proc != nil {
		return p.proc.GetOutType()
	} else if p.arr != nil {
		return reflect.TypeOf(p.arr).Elem()
	} else {
		panic("both proc and arr are nil")
	}
}

func (p *Pipe) ToSlice() interface{} {
	if p.proc == nil {
		return p.arr
	}
	length := p.srcLen()
	outElemType := p.getOutType()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	for i := 0; i < length; i++ {
		itemValue, keep := p.getValue(i)
		if keep {
			newSliceValue = reflect.Append(newSliceValue, itemValue)
		}
	}
	return newSliceValue.Interface()
}

func (p *Pipe) ToMap(getKey, getVal interface{}) interface{} {
	getKeyValue := reflect.ValueOf(getKey)
	getKeyType := getKeyValue.Type()
	getValValue := reflect.ValueOf(getVal)
	getValType := getValValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
		panic("getKey func invalid")
	}
	if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
		panic("getVal func invalid")
	}
	keyType := getKeyType.Out(0)
	valType := getValType.Out(0)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, valType))
	length := p.srcLen()
	for i := 0; i < length; i++ {
		itemValue, keep := p.getValue(i)
		if keep {
			newMapValue.SetMapIndex(
				getKeyValue.Call([]reflect.Value{itemValue})[0],
				getValValue.Call([]reflect.Value{itemValue})[0],
			)
		}
	}
	return newMapValue.Interface()
}

func (p *Pipe) ToGroupMap(getKey, getVal interface{}) interface{} {
	getKeyValue := reflect.ValueOf(getKey)
	getKeyType := getKeyValue.Type()
	getValValue := reflect.ValueOf(getVal)
	getValType := getValValue.Type()
	outElemType := p.getOutType()
	if !isGoodFunc(getKeyType, []interface{}{outElemType}, []interface{}{nil}) {
		panic("getKey func invalid")
	}
	if !isGoodFunc(getValType, []interface{}{outElemType}, []interface{}{nil}) {
		panic("getVal func invalid")
	}
	keyType := getKeyType.Out(0)
	valType := getValType.Out(0)
	sliceType := reflect.SliceOf(valType)
	newMapValue := reflect.MakeMap(reflect.MapOf(keyType, sliceType))
	//zeroValue := reflect.Zero(sliceType)
	length := p.srcLen()
	for i := 0; i < length; i++ {
		itemValue, keep := p.getValue(i)
		if !keep {
			continue
		}
		keyValue := getKeyValue.Call([]reflect.Value{itemValue})[0]
		valValue := getValValue.Call([]reflect.Value{itemValue})[0]
		slot := newMapValue.MapIndex(keyValue)
		if !slot.IsValid() {
			slot = reflect.MakeSlice(sliceType, 0, length-i)
		}
		slot = reflect.Append(slot, valValue)
		newMapValue.SetMapIndex(keyValue, slot)
	}
	return newMapValue.Interface()
}

func (p *Pipe) Reduce(initValue interface{}, proc interface{}) interface{} {
	length := p.srcLen()
	outElemType := p.getOutType()
	procValue := reflect.ValueOf(proc)
	procType := procValue.Type()
	initType := reflect.TypeOf(initValue)
	if !isGoodFunc(procType, []interface{}{initType, outElemType}, []interface{}{initType}) {
		panic("reduce function invalid")
	}
	for i := 0; i < length; i++ {
		itemValue, keep := p.getValue(i)
		if keep {
			outs := procValue.Call([]reflect.Value{reflect.ValueOf(initValue), itemValue})
			initValue = outs[0].Interface()
		}
	}
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

func (p *Pipe) Sort(less interface{}) *Pipe {
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
	return &Pipe{
		arr: delegate.Arr.Interface(),
	}
}

func (p *Pipe) Reverse() *Pipe {
	outElemType := p.getOutType()
	length := p.srcLen()
	newSliceType := reflect.SliceOf(outElemType)
	newSliceValue := reflect.MakeSlice(newSliceType, 0, length)
	for i := length - 1; i >= 0; i-- {
		itemValue, keep := p.getValue(i)
		if keep {
			newSliceValue = reflect.Append(newSliceValue, itemValue)
		}
	}
	return &Pipe{
		arr: newSliceValue.Interface(),
	}
}
