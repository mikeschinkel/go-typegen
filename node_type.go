package typegen

import (
	"reflect"
)

type NodeType uint

const (
	PointerNode   = NodeType(reflect.Pointer)
	MapNode       = NodeType(reflect.Map)
	ArrayNode     = NodeType(reflect.Array)
	SliceNode     = NodeType(reflect.Slice)
	StructNode    = NodeType(reflect.Struct)
	InterfaceNode = NodeType(reflect.Interface)
	StringNode    = NodeType(reflect.String)
	IntNode       = NodeType(reflect.Int)
	Int8Node      = NodeType(reflect.Int8)
	Int16Node     = NodeType(reflect.Int16)
	Int32Node     = NodeType(reflect.Int32)
	Int64Node     = NodeType(reflect.Int64)
	UintNode      = NodeType(reflect.Uint)
	Uint8Node     = NodeType(reflect.Uint8)
	Uint16Node    = NodeType(reflect.Uint16)
	Uint32Node    = NodeType(reflect.Uint32)
	Uint64Node    = NodeType(reflect.Uint64)
	Float32Node   = NodeType(reflect.Float32)
	Float64Node   = NodeType(reflect.Float64)
	BoolNode      = NodeType(reflect.Bool)
	InvalidNode   = NodeType(reflect.Invalid)
	RefNode       = NodeType(reflect.UnsafePointer + 10)
	FieldNode     = NodeType(reflect.UnsafePointer + 11)
	ElementNode   = NodeType(reflect.UnsafePointer + 12)
)

func (nt NodeType) String() string {
	return nodeTypeName(nt)
}

func nodeTypeName(nt NodeType) (s string) {
	switch nt {
	case PointerNode:
		s = "Pointer"
	case MapNode:
		s = "Map"
	case ArrayNode:
		s = "Array"
	case SliceNode:
		s = "Slice"
	case StructNode:
		s = "Struct"
	case InterfaceNode:
		s = "Interface"
	case StringNode:
		s = "String"
	case IntNode:
		s = "Int"
	case Int8Node:
		s = "Int8"
	case Int16Node:
		s = "Int16"
	case Int32Node:
		s = "Int32"
	case Int64Node:
		s = "Int64"
	case UintNode:
		s = "UInt"
	case Uint8Node:
		s = "Uint8"
	case Uint16Node:
		s = "Uint16"
	case Uint32Node:
		s = "Uint32"
	case Uint64Node:
		s = "Uint64"
	case Float32Node:
		s = "Float32"
	case Float64Node:
		s = "Float64"
	case BoolNode:
		s = "Bool"
	case InvalidNode:
		s = "Invalid"
	case RefNode:
		s = "Ref"
	case FieldNode:
		s = "Field"
	case ElementNode:
		s = "Element"
	default:
		panicf("Invalid node type: %d", nt)
	}
	return s
}
