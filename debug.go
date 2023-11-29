//go:build debug

package typegen

import (
	"fmt"
	"strings"
)

func init() {
	(&Node{}).DebugString()
	(&NodeMarshaler{}).DebugString()
	resetDebugString = func(a any) {
		switch t := a.(type) {
		case *Node:
			t.debugString = fmt.Sprintf("%s %sNode [Id: %d, Index: %d]", t.Name, t.Type, t.Id, t.Index)
			return
		case *NodeMarshaler:
			sb := strings.Builder{}
			for index := len(t.nodes) - 1; index >= 1; index-- {
				sb.WriteByte(' ')
				sb.WriteString(t.nodes[index].Type.String())
			}
			t.debugString = fmt.Sprintf("[%d]%s", len(t.nodeMap), sb.String())
		}
	}
}

func (n *Node) DebugString() string {
	return n.debugString
}

func (m *NodeMarshaler) DebugString() string {
	return m.debugString
}
