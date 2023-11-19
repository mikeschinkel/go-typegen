package typegen

import (
	"fmt"
	"reflect"
)

type Nodes []*Node

type NodeArgs struct {
	Name      string
	NodeRef   *Node
	Type      NodeType
	marshaler *NodeMarshaler
	Value     reflect.Value
	Index     int
	Parent    *Node
}

type Node struct {
	Value       reflect.Value
	NodeRef     *Node
	Type        NodeType
	Name        string
	parent      *Node
	nodes       Nodes
	marshaler   *NodeMarshaler
	Indent      string
	Index       int
	varname     string
	debugString string
}

func NewNode(args *NodeArgs) (n *Node) {
	if args.Type == InvalidNode {
		args.Type = NodeType(args.Value.Kind())
	}
	n = &Node{
		Name:      args.Name,
		Type:      args.Type,
		NodeRef:   args.NodeRef,
		nodes:     make(Nodes, 0),
		marshaler: args.marshaler,
		Value:     args.Value,
		Index:     args.Index,
		parent:    args.Parent,
	}
	n.resetDebugString()
	return n
}

func (n *Node) String() string {
	n.resetDebugString()
	return n.debugString
}

func (n *Node) resetDebugString() {
	n.debugString = fmt.Sprintf("%s %sNode [Index: %d]", n.Name, n.Type, n.Index)
}

func (n *Node) Varname() string {
	return n.varname
}

func (n *Node) SetVarname(name string) {
	if n.varname != "" {
		panicf("Unexpected: overwriting varname '%s'", name)
	}
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
