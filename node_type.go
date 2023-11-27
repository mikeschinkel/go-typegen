package typegen

import (
	"reflect"

	. "github.com/mikeschinkel/go-lib"
)

type NodeType uint

const (
	PointerNode       = NodeType(reflect.Pointer)
	MapNode           = NodeType(reflect.Map)
	ArrayNode         = NodeType(reflect.Array)
	SliceNode         = NodeType(reflect.Slice)
	StructNode        = NodeType(reflect.Struct)
	InterfaceNode     = NodeType(reflect.Interface)
	StringNode        = NodeType(reflect.String)
	IntNode           = NodeType(reflect.Int)
	Int8Node          = NodeType(reflect.Int8)
	Int16Node         = NodeType(reflect.Int16)
	Int32Node         = NodeType(reflect.Int32)
	Int64Node         = NodeType(reflect.Int64)
	UintptrNode       = NodeType(reflect.Uintptr)
	UintNode          = NodeType(reflect.Uint)
	Uint8Node         = NodeType(reflect.Uint8)
	Uint16Node        = NodeType(reflect.Uint16)
	Uint32Node        = NodeType(reflect.Uint32)
	Uint64Node        = NodeType(reflect.Uint64)
	Float32Node       = NodeType(reflect.Float32)
	Float64Node       = NodeType(reflect.Float64)
	BoolNode          = NodeType(reflect.Bool)
	FuncNode          = NodeType(reflect.Func)
	InvalidNode       = NodeType(reflect.Invalid)
	UnsafePointerNode = NodeType(reflect.UnsafePointer)
	FieldNode         = NodeType(reflect.UnsafePointer + 10)
	ElementNode       = NodeType(reflect.UnsafePointer + 11)
	SubstitutionNode  = NodeType(reflect.UnsafePointer + 12)
)

var (
	// ScalarNodeTypes groups NodeTypes that represent scalar values into a slice for
	// convenient use with isOneOf() which is called in .scalarChildWritten().
	ScalarNodeTypes = []NodeType{
		StringNode,
		IntNode,
		Int8Node,
		Int16Node,
		Int32Node,
		Int64Node,
		UintptrNode,
		UintNode,
		Uint8Node,
		Uint16Node,
		Uint32Node,
		Uint64Node,
		Float32Node,
		Float64Node,
		BoolNode,
		UnsafePointerNode,
		SubstitutionNode,
	}
)

func (nt NodeType) String() string {
	return nodeTypeName(nt)
}

func nodeTypeName(nt NodeType) (s string) {
	switch nt {
	case PointerNode:
		s = "pointer"
	case MapNode:
		s = "map"
	case ArrayNode:
		s = "array"
	case SliceNode:
		s = "slice"
	case StructNode:
		s = "struct"
	case InterfaceNode:
		s = "interface"
	case StringNode:
		s = "string"
	case IntNode:
		s = "int"
	case Int8Node:
		s = "int8"
	case Int16Node:
		s = "int16"
	case Int32Node:
		s = "int32"
	case Int64Node:
		s = "int64"
	case UintNode:
		s = "uInt"
	case Uint8Node:
		s = "uint8"
	case Uint16Node:
		s = "uint16"
	case Uint32Node:
		s = "uint32"
	case Uint64Node:
		s = "uint64"
	case Float32Node:
		s = "float32"
	case Float64Node:
		s = "float64"
	case BoolNode:
		s = "bool"
	case FuncNode:
		s = "func"
	case InvalidNode:
		s = "invalid"
	case FieldNode:
		s = "field"
	case ElementNode:
		s = "element"
	case UnsafePointerNode:
		s = "unsafePointer"
	case UintptrNode:
		s = "uintptr"
	case SubstitutionNode:
		s = "substitution"
	default:
		Panicf("Invalid node type: %d", nt)
	}
	return s
}
