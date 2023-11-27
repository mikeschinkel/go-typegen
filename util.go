package typegen

import (
	"fmt"
	"regexp"
	"strings"
)

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

// resetDebugString provides a stub for replacement by debug.go for when the
// `debug` tag is used.
var resetDebugString = func(any) {}

var iFaceRE = regexp.MustCompile(`^(\W*)interface \{}`)

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
