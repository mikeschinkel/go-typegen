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
