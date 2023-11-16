package typegen

func init() {
	(&Node{}).DebugString()
}

func (n *Node) DebugString() string {
	return n.debugString
}
