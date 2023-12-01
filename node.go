package typegen

import (
	"fmt"
	"reflect"

	. "github.com/mikeschinkel/go-lib"
	"github.com/mikeschinkel/go-typegen/ezreflect"
)

type Nodes []*Node

func (ns Nodes) AppendNode(n *Node) Nodes {
	resetDebugString(n)
	return append(ns, n)
}

type NodeArgs struct {
	Value        any
	ReflectValue *reflect.Value
	Name         string
	Type         NodeType
	marshaler    *NodeMarshaler
	Index        int
	Parent       *Node
	Typename     string
}

type Node struct {
	Id          int
	Value       any
	Typename    string
	Type        NodeType
	Name        string
	Parent      *Node
	nodes       Nodes
	Marshaler   *NodeMarshaler
	Index       int
	varname     string
	debugString string
}

func NewNode(id int, args *NodeArgs) (n *Node) {

	if args.Type == InvalidNode {
		args.Type = NodeType(args.ReflectValue.Kind())
	}

	n = &Node{
		Id:        id,
		Name:      args.Name,
		Type:      args.Type,
		Typename:  args.Typename,
		Value:     args.Value,
		Marshaler: args.marshaler,
		Index:     args.Index,
		Parent:    args.Parent,
	}
	if args.ReflectValue != nil {
		n.Value = ezreflect.AsAny(*args.ReflectValue)
		n.Typename = ezreflect.TypenameOf(*args.ReflectValue)
	}

	if n.Typename == "" {
		Panicf("Missing argument for NewNode() for '%s'; either ReflectValue or Typename must be passed.", n)
	}

	n.Reset()

	return n
}

func (n *Node) String() (s string) {
	if n != nil {
		s = fmt.Sprintf("%s %sNode [Id: %d, Index: %d]", n.Name, n.Type, n.Id, n.Index)
	}
	return s
}

func (n *Node) Reset() *Node {
	if n == nil {
		return n
	}
	n.nodes = make(Nodes, 0)
	resetDebugString(n)
	return n
}

func (n *Node) Varname() string {
	return n.varname
}

func (n *Node) SetVarname(name string) *Node {
	if n.varname != "" {
		Panicf("Unexpected: overwriting varname '%s'", name)
	}
	n.varname = name
	return n
}

func (n *Node) SetNodeCount(cnt int) *Node {
	nodes := make(Nodes, 0, cnt)
	for i, node := range n.nodes {
		if i >= len(nodes) {
			goto end
		}
		nodes[i] = node
	}
end:
	n.nodes = nodes
	return n
}

func (n *Node) AddNode(node *Node) *Node {
	node.Parent = n
	resetDebugString(node)
	node.Index = len(n.nodes)
	n.nodes = n.nodes.AppendNode(node)

	return n
}

func (n *Node) ChildNode(i int) *Node {
	if i < 0 {
		n = nil
		goto end
	}
	if i >= len(n.nodes) {
		n = nil
		goto end
	}
	n = n.nodes[i]
end:
	return n
}

func (n *Node) Nodes() Nodes {
	return n.nodes
}
