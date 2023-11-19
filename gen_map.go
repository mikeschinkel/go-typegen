package typegen

import (
	"reflect"
)

// GenMap is a map type for containing the Nodes that have been generated so that
// we can avoid generating a Node multiple times. It is keyed by value of
// Node.Value and its value will be the corresponding *Node. Used as a property
// in CodeBuilder.
type GenMap map[reflect.Value]*Node
