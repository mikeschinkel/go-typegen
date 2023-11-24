package ezreflect

import (
	"fmt"
	"reflect"
	"strconv"
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
	if !rv.IsValid() {
		s = "nil"
		goto end
	}
	switch rv.Kind() {
	case reflect.Interface:
		// TODO: What about named interfaces?
		s = fmt.Sprintf("any(%s)", AsString(ChildOf(rv)))
	case reflect.Pointer:
		// TODO: This is probably wrong
		s = fmt.Sprintf("*%s", AsString(ChildOf(rv)))
	case reflect.String:
		s = strconv.Quote(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16:
		s = strconv.Itoa(int(rv.Int()))
	case reflect.Int32, reflect.Int64:
		s = strconv.FormatInt(rv.Int(), 10)
	case reflect.Float32:
		s = strconv.FormatFloat(rv.Float(), 'g', 10, 32)
	case reflect.Float64:
		s = strconv.FormatFloat(rv.Float(), 'g', 10, 64)
	case reflect.Map, reflect.Slice, reflect.Struct:
		s = fmt.Sprintf("%s{...}", TypenameOf(rv))
	case reflect.Bool:
		if rv.Bool() {
			s = "true"
		} else {
			s = "false"
		}
	default:
		panicf("Unhandled (s of yet) reflect value kind: %s", rv.Kind())
	}
end:
	return s
}

func TypenameOf(rv reflect.Value) (n string) {
	var rt reflect.Type

	if !rv.IsValid() {
		n = "nil"
		goto end
	}
	rt = rv.Type()
	switch rv.Kind() {
	case reflect.Interface:
		n = fmt.Sprintf("any(%s)", TypenameOf(ChildOf(rv)))
	case reflect.Pointer:
		n = fmt.Sprintf("*%s", TypenameOf(ChildOf(rv)))
	default:
		n = rt.String()
	}
end:
	return n
}

func ChildOf(rv reflect.Value) (c reflect.Value) {
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		c = rv.Elem()
	}
	return c
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
