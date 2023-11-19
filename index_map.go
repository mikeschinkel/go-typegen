package typegen

import (
	"reflect"
)

// IndexMap is used for a Node lookup nap keyed by reflect.Value with index into
// .nodes for it value. Used to find a node to nullify in .nodes if it does not
// need to be generated. See .scalarChildWritten to see if used. Used as a
// property in CodeBuilder.
type IndexMap map[reflect.Value]int
