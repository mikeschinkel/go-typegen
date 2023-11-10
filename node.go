package typegen

import (
	"fmt"
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
	IntNode        = NodeType(reflect.Int64)
	UIntNode       = NodeType(reflect.Uint64)
	FloatNode      = NodeType(reflect.Float64)
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
	Type        NodeType
	CodeBuilder *CodeBuilder
	Value       reflect.Value
	Index       int
}

type Node struct {
	Value       reflect.Value
	Ref         reflect.Value
	Type        NodeType
	Name        string
	parent      *Node
	nodes       Nodes
	codeBuilder *CodeBuilder
	Indent      string
	Index       int
}

func NewNode(args *NodeArgs) *Node {
	if args.Type == InvalidNode {
		args.Type = getNodeType(args.Value.Kind())
	}
	return &Node{
		Name:        args.Name,
		Type:        args.Type,
		nodes:       make(Nodes, 0),
		codeBuilder: args.CodeBuilder,
		Value:       args.Value,
		Index:       args.Index,
	}
}

func (n *Node) Varname() string {
	return fmt.Sprintf("var%d", n.Index)
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

func getNodeType(rk reflect.Kind) (nt NodeType) {
	// Start with the type definition
	switch rk {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		nt = IntNode
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		nt = UIntNode
	case reflect.Float32, reflect.Float64:
		nt = FloatNode
	case reflect.Bool:
		nt = BoolNode
	case reflect.Invalid:
		nt = InvalidNode
	default:
		nt = NodeType(rk)
	}
	return nt
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
	case UIntNode:
		s = "UInt"
	case FloatNode:
		s = "Float"
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
