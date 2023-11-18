package typegen

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
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

var iFaceRE = regexp.MustCompile(`^(\W*)interface \{\}`)

func replaceInterfaceWithAny(name string) string {
	return iFaceRE.ReplaceAllString(name, "${1}any")
}

// maybeStripPackage will remove `foo.` from `foo.Bar`, *foo.Bar`, []foo.Bar` and so on.
func maybeStripPackage(name, omitPkg string) string {
	var pkgStripRE *regexp.Regexp

	if name == "&" {
		goto end
	}
	if len(name) == 0 {
		goto end
	}
	if !strings.Contains(name, ".") {
		goto end
	}
	pkgStripRE = regexp.MustCompile(fmt.Sprintf(`^(\W*)%s\.`, omitPkg))
	name = pkgStripRE.ReplaceAllString(name, "$1")
end:
	return name
}

func isOneOf[T comparable](val T, options ...T) bool {
	for _, option := range options {
		if val == option {
			return true
		}
	}
	return false
}
