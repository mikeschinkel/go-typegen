package easyreflect

import (
	"fmt"
	"reflect"
)

type Reflector struct {
	Value          any
	ReflectValue   reflect.Value
	ReflectType    reflect.Type
	ValueReflector *Reflector
}

// ContainedValue returns the Reflector for the value contained within an interface
func (r *Reflector) ContainedValue() (_ *Reflector, err error) {
	if r.ReflectValue.Kind() != reflect.Interface {
		err = fmt.Errorf("value not an interface; %#v", r.Value)
		goto end
	}
	if r.ReflectValue.IsNil() {
		err = fmt.Errorf("interface value is nil")
		goto end
	}
	r = NewReflector(r.ReflectValue.Elem())
end:
	return r, err
}

// DereferencedValue returns the Reflector for the value contained within an interface
func (r *Reflector) DereferencedValue() (_ *Reflector, err error) {
	if r.ReflectValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("value not a pointer; %#v", r.Value)
		goto end
	}
	if r.ReflectValue.IsNil() {
		err = fmt.Errorf("pointer value is nil")
		goto end
	}
	r = NewReflector(r.ReflectValue.Elem())
end:
	return r, err
}

func NewReflector(value any) *Reflector {
	return &Reflector{
		Value:        value,
		ReflectValue: reflect.ValueOf(value),
		ReflectType:  reflect.TypeOf(value),
	}
}
