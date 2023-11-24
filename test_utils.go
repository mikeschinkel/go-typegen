//go:build test

package typegen

func TestInitNode(n *Node) *Node {
	if n == nil {
		return nil
	}
	n.nodes = make(Nodes, 0)
	return n
}
func TestAddNode(parent, child *Node) *Node {
	TestInitNode(child)
	child.ResetDebugString()
	parent.nodes = append(parent.nodes, child)
	return parent.ResetDebugString()
}
func TestGetNode(n *Node, i int) *Node {
	return n.nodes[i]
}

func TestFixupNodes(ns Nodes, f func(Nodes)) Nodes {
	for i, n := range ns {
		if n == nil {
			continue
		}
		ns[i] = n.Reset()
	}
	if f != nil {
		f(ns)
	}
	return ns
}
