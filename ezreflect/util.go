package ezreflect

import (
	"fmt"
	"reflect"

	"github.com/mikeschinkel/go-diffator"
)

func panicf(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}

func AsAny(rv reflect.Value) (a any) {
	if !rv.IsValid() {
		a = nil
		goto end
	}
	if rv.CanInterface() {
		a = rv.Interface()
		goto end
	}
	switch rv.Kind() {
	case reflect.Bool:
		a = rv.Bool()
	case reflect.String:
		a = rv.String()
	case reflect.Int:
		a = int(rv.Int())
	case reflect.Int8:
		a = int8(rv.Int())
	case reflect.Int16:
		a = int16(rv.Int())
	case reflect.Int32:
		a = int32(rv.Int())
	case reflect.Int64:
		a = rv.Int()
	case reflect.Float32:
		a = float32(rv.Float())
	case reflect.Float64:
		a = rv.Float()
	case reflect.Map, reflect.Slice:
		a = TypenameOf(rv)
	case reflect.Interface, reflect.Pointer:
		a = AsAny(ChildOf(rv))
	default:
		panicf("Unhandled reflect value kind (as of yet): %s", rv.Kind())
	}
end:
	return a
}

func AsString(rv reflect.Value) (s string) {
	return diffator.NewReflector().AsString(rv)
}

func TypenameOf(rv reflect.Value) (n string) {
	return diffator.NewReflector().TypenameOf(rv)
}

func ChildOf(rv reflect.Value) (c reflect.Value) {
	return diffator.NewReflector().ChildOf(rv)
}

func Equivalent(v1, v2 any) (same bool) {
	rv1 := reflect.ValueOf(v1)
	rv2 := reflect.ValueOf(v2)
	if rv1.Kind() == reflect.Pointer {
		rv1 = rv1.Elem()
	}
	if rv1.Kind() != rv2.Kind() {
		goto end
	}
	if rv1.Comparable() && !rv1.Equal(rv2) {
		goto end
	}
	if !reflect.DeepEqual(rv1, rv2) {
		goto end
	}
	same = true
end:
	return same
}
