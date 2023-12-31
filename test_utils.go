//go:build test

package typegen

import (
	"fmt"
	"os"
)

func init() {
	fmt.Fprintln(os.Stderr, "Testing...")
}

func TestInitNodes(nodes Nodes) Nodes {
	for i, n := range nodes {
		nodes[i] = TestInitNode(n)
	}
	return nodes
}
func TestInitNode(n *Node) *Node {
	if n == nil {
		goto end
	}
	if n.nodes != nil {
		goto end
	}
	n.nodes = make(Nodes, 0)
end:
	return n
}

func TestAddNode(parent, child *Node) *Node {
	TestInitNode(child)
	resetDebugString(child)
	parent.nodes = append(parent.nodes, child)
	resetDebugString(parent)
	return parent
}

func TestResetNode(n *Node) *Node {
	resetDebugString(n)
	return n
}

func TestGetNode(n *Node, i int) *Node {
	return n.nodes[i]
}

func TestGetNodes(n *Node) Nodes {
	return n.nodes
}

func TestFixupNodes(ns Nodes, f func(Nodes)) Nodes {
	for i, n := range ns {
		if n == nil {
			continue
		}
		ns[i] = n.Reset()
	}
	if f != nil {
		f(TestInitNodes(ns))
	}
	return ns
}
