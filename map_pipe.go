package pipe

import (
	"reflect"
)

type MapPipe struct {
	m interface{}
	t reflect.Type
	v reflect.Value
}

func NewMapPipe(m interface{}) *MapPipe {
	mValue := reflect.ValueOf(m)
	mType := mValue.Type()
	if mType.Kind() != reflect.Map {
		panic("NewMapPipe only accept map param")
	}
	return &MapPipe{
		m: m,
		v: mValue,
		t: mType,
	}
}

func (mp *MapPipe) Keys() *Pipe {
	elType := mp.t.Key()
	sliceType := reflect.SliceOf(elType)
	keysValue := mp.v.MapKeys()
	length := len(keysValue)
	newSliceValue := reflect.MakeSlice(sliceType, 0, length)
	for _, keyValue := range keysValue {
		newSliceValue = reflect.Append(newSliceValue, keyValue)
	}
	return NewPipe(newSliceValue.Interface())
}

func (mp *MapPipe) Values() *Pipe {
	elType := mp.t.Elem()
	sliceType := reflect.SliceOf(elType)
	keysValue := mp.v.MapKeys()
	length := len(keysValue)
	newSliceValue := reflect.MakeSlice(sliceType, 0, length)
	for _, keyValue := range keysValue {
		newSliceValue = reflect.Append(newSliceValue, mp.v.MapIndex(keyValue))
	}
	return NewPipe(newSliceValue.Interface())
}
