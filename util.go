package typegen

import (
	"fmt"
	"reflect"
)

func panicf(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}

func filterMapFunc[M ~map[K]V, K comparable, V any](m M, match func(K, V) bool) M {
	f := make(M, len(m))
	for k, v := range m {
		if !match(k, v) {
			continue
		}
		f[k] = v
	}
	return f
}

//// setReflectValueToNil attempts to nil the field represented by 'fieldValue'
//func setReflectValueToNil(rv reflect.Value) {
//	var ptr unsafe.Pointer
//	var ifacePtr *interface{}
//
//	// rv must be addressable to use unsafe to set it
//	if rv.CanSet() {
//		// The field is exported, and we can set it safely.
//		rv.Set(reflect.Zero(rv.Type()))
//		goto end
//	}
//	if !rv.CanAddr() {
//		goto end
//	}
//
//	// If it's not addressable, we can't change it directly, so we need to use unsafe.
//	// Obtain the address of the field as a pointer (uintptr).
//	ptr = unsafe.Pointer(rv.UnsafeAddr())
//
//	// Convert the field pointer to a pointer to an empty interface.
//	ifacePtr = (*interface{})(ptr)
//
//	// Set the field to nil using the empty interface pointer.
//	*ifacePtr = nil
//end:
//}

func isSame(v1, v2 reflect.Value) (same bool) {
	if v1.Kind() == reflect.Pointer {
		v1 = v1.Elem()
	}
	if v1.Kind() != v2.Kind() {
		goto end
	}
	if v1.Comparable() && !v1.Equal(v2) {
		goto end
	}
	if !reflect.DeepEqual(v1, v2) {
		goto end
	}
	same = true
end:
	return same
}
