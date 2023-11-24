package ezreflect

import (
	"reflect"
)

type Reflector struct {
	value        any
	reflectValue reflect.Value
	reflectType  reflect.Type
}

type ReflectorArgs struct {
	Value        any
	ReflectValue reflect.Value
	ReflectType  reflect.Type
}

func NewReflector(value any) *Reflector {
	return &Reflector{
		value:        value,
		reflectValue: reflect.ValueOf(value),
		reflectType:  reflect.TypeOf(value),
	}
}
func NewReflectWrapper(rv reflect.Value) *Reflector {
	return NewReflectorFromArgs(&ReflectorArgs{
		ReflectValue: rv,
	})
}
func NewReflectorFromArgs(args *ReflectorArgs) *Reflector {
	r := &Reflector{
		value:        args.Value,
		reflectValue: args.ReflectValue,
		reflectType:  args.ReflectType,
	}
	if !r.reflectValue.IsValid() {
		r.reflectValue = reflect.ValueOf(r.value)
	}
	if r.reflectType == nil {
		if r.reflectValue.IsValid() {
			r.reflectType = r.reflectValue.Type()
		}
	}
	return r
}

func (r *Reflector) Child() (c *Reflector) {
	return NewReflectWrapper(ChildOf(r.reflectValue))
}

func (r *Reflector) Typename() (s string) {
	return TypenameOf(r.reflectValue)
}

func (r *Reflector) Any() (a any) {
	return AsAny(r.reflectValue)
}
func (r *Reflector) String() (s string) {
	return AsString(r.reflectValue)
}
