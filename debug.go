//go:build debug

package typegen

func init() {
	resetDebugString = func(a any) {
		switch t := a.(type) {
		case *Node:
			t.debugString = t.String()
		case *NodeMarshaler:
			t.debugString = t.String()
		}
	}
	(&Node{}).DebugString()
	(&NodeMarshaler{}).DebugString()
}

func (n *Node) DebugString() string {
	return n.debugString
}

func (m *NodeMarshaler) DebugString() string {
	return m.debugString
}
