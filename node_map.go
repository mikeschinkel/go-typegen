package typegen

import (
	"reflect"
)

type NodeMap map[reflect.Value]*Node

func (m NodeMap) Len() int {
	return len(m)
}
