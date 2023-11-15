package typegen

import (
	"reflect"
)

type NodeType uint

const (
	PointerNode    = NodeType(reflect.Pointer)
	MapNode        = NodeType(reflect.Map)
	ArrayNode      = NodeType(reflect.Array)
	SliceNode      = NodeType(reflect.Slice)
	StructNode     = NodeType(reflect.Struct)
	InterfaceNode  = NodeType(reflect.Interface)
	StringNode     = NodeType(reflect.String)
	IntNode        = NodeType(reflect.Int)
	Int8Node       = NodeType(reflect.Int8)
	Int16Node      = NodeType(reflect.Int16)
	Int32Node      = NodeType(reflect.Int32)
	Int64Node      = NodeType(reflect.Int64)
	UintNode       = NodeType(reflect.Uint)
	Uint8Node      = NodeType(reflect.Uint8)
	Uint16Node     = NodeType(reflect.Uint16)
	Uint32Node     = NodeType(reflect.Uint32)
	Uint64Node     = NodeType(reflect.Uint64)
	Float32Node    = NodeType(reflect.Float32)
	Float64Node    = NodeType(reflect.Float64)
	BoolNode       = NodeType(reflect.Bool)
	InvalidNode    = NodeType(reflect.Invalid)
	RefNode        = NodeType(reflect.UnsafePointer + 10)
	FieldNameNode  = NodeType(reflect.UnsafePointer + 11)
	FieldValueNode = NodeType(reflect.UnsafePointer + 12)
	MapKeyNode     = NodeType(reflect.UnsafePointer + 13)
	MapValueNode   = NodeType(reflect.UnsafePointer + 14)
	ElementNode    = NodeType(reflect.UnsafePointer + 15)
)

type Nodes []*Node

func (ns Nodes) Len() int {
	return len(ns)
}

func (ns Nodes) WriteCode(g *Generator) {
	n := 0
	for _, node := range ns {
		node.Index = n
		g.WriteCode(node)
		n++
	}
}

type NodeArgs struct {
	Name        string
	NodeRef     *Node
	Type        NodeType
	CodeBuilder *CodeBuilder
	Value       reflect.Value
	Index       int
}

type Node struct {
	Value       reflect.Value
	Ref         reflect.Value
	NodeRef     *Node
	Type        NodeType
	Name        string
	parent      *Node
	nodes       Nodes
	codeBuilder *CodeBuilder
	Indent      string
	Index       int
	varname     string
}

func NewNode(args *NodeArgs) *Node {
	if args.Type == InvalidNode {
		args.Type = NodeType(args.Value.Kind())
	}
	return &Node{
		Name:        args.Name,
		Type:        args.Type,
		NodeRef:     args.NodeRef,
		nodes:       make(Nodes, 0),
		codeBuilder: args.CodeBuilder,
		Value:       args.Value,
		Index:       args.Index,
	}
}

func (n *Node) Varname() string {
	return n.varname
}

func (n *Node) SetVarname(name string) {
	n.varname = name
}

func (n *Node) SetNodeCount(cnt int) {
	nodes := make(Nodes, 0, cnt)
	for i, node := range n.nodes {
		if i >= len(nodes) {
			goto end
		}
		nodes[i] = node
	}
end:
	n.nodes = nodes
}

func (n *Node) AddNode(node *Node) *Node {
	node.parent = n
	n.nodes = append(n.nodes, node)
	return n
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
	case FieldNameNode:
		s = "FieldName"
	case FieldValueNode:
		s = "FieldValue"
	case MapKeyNode:
		s = "MapKey"
	case MapValueNode:
		s = "MapValue"
	case ElementNode:
		s = "Element"
	default:
		panicf("Invalid node type: %d", nt)
	}
	return s
}
