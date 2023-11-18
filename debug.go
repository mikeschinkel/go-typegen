package typegen

func init() {
	(&Node{}).DebugString()
	(&NodeMarshaler{}).DebugString()
}

func (n *Node) DebugString() string {
	return n.debugString
}

func (m *NodeMarshaler) DebugString() string {
	return m.debugString
}
