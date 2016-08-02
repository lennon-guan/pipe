package pipe

import (
	"fmt"
	"reflect"
)

type IProc interface{
	Next(reflect.Value) (reflect.Value, bool)
	GetOutType() reflect.Type
}

type MapProc struct {
	InType reflect.Type
	OutType reflect.Type
	Func reflect.Value
}

func NewMapProc(f interface{}) *MapProc {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if fType.Kind() != reflect.Func {
		panic("map argument must be a function")
	}
	if fType.NumIn() != 1 || fType.NumOut() != 1 {
		panic("map function must has only one input parameter and only one output parameter")
	}
	return &MapProc {
		InType: fType.In(0),
		OutType: fType.Out(0),
		Func: fValue,
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
	InType reflect.Type
	OutType reflect.Type
	Func reflect.Value
}

func NewFilterProc(f interface{}) *FilterProc {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if fType.Kind() != reflect.Func {
		panic("filter argument must be a function")
	}
	if fType.NumIn() != 1 || fType.NumOut() != 1 || fType.Out(0).Kind() != reflect.Bool {
		panic("filter function must has only one input parameter and only one boolean output parameter")
	}
	return &FilterProc {
		InType: fType.In(0),
		OutType: fType.In(0),
		Func: fValue,
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
	arr interface{}
	srcPipe *Pipe
	proc IProc
}

func NewPipe(arr interface{}) *Pipe {
	return &Pipe{
		arr: arr,
		srcPipe: nil,
		proc: nil,
	}
}

func (p *Pipe) Filter(proc interface{}) *Pipe {
	return &Pipe {
		arr: nil,
		srcPipe: p,
		proc: NewFilterProc(proc),
	}
}

func (p *Pipe) Map(proc interface{}) *Pipe {
	return &Pipe {
		arr: nil,
		srcPipe: p,
		proc: NewMapProc(proc),
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

func (p *Pipe) ToSlice() interface{} {
	if p.proc == nil {
		return p.arr
	}
	length := p.srcLen()
	outElemType := p.proc.GetOutType()
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

func (p *Pipe) Reduce(initValue interface{}, proc interface{}) interface{} {
	length := p.srcLen()
	var outElemType reflect.Type
	if p.proc != nil {
		outElemType = p.proc.GetOutType()
	} else if p.arr != nil {
		outElemType = reflect.TypeOf(p.arr).Elem()
	} else {
		panic("both proc and arr are nil")
	}
	procValue := reflect.ValueOf(proc)
	procType := procValue.Type()
	initType := reflect.TypeOf(initValue)
	if procType.Kind() != reflect.Func || procType.NumIn() != 2 ||
		procType.In(0) != reflect.TypeOf(initValue) || procType.In(1) != outElemType ||
		procType.NumOut() != 1 || procType.Out(0) != initType {
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

