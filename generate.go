package typegen

import (
	"fmt"
	"strconv"
	"strings"
)

type Generator struct {
	strings.Builder
	Indent string
	Depth  int
}

func NewGenerator() *Generator {
	return &Generator{
		Indent: "  ",
	}
}

func (g *Generator) WriteCode(n *Node) {
	switch n.Type {
	case RefNode:
		g.RefNode(n)
	case StructNode:
		g.StructNode(n)
	case PointerNode:
		g.PointerNode(n)
	case MapNode:
		g.MapNode(n)
	case ArrayNode:
		g.ArrayNode(n)
	case InterfaceNode:
		g.InterfaceNode(n)
	case StringNode:
		g.StringNode(n)
	case IntNode:
		g.IntNode(n)
	case UIntNode:
		g.UIntNode(n)
	case FloatNode:
		g.FloatNode(n)
	case BoolNode:
		g.BoolNode(n)
	case InvalidNode:
		g.InvalidNode(n)
	case FieldNameNode:
		g.FieldNameNode(n)
	case FieldValueNode:
		g.FieldValueNode(n)
	case ElementNode:
		g.ElementNode(n)
	//case MapKeyNode:
	//	g.MapKeyNode(n)
	//case MapValueNode:
	//	g.MapValueNode(n)

	case SliceNode:
		g.SliceNode(n)
	}
}

func (g *Generator) ArrayNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) InterfaceNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) InvalidNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) FieldNameNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) FieldValueNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) ElementNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) RefNode(n *Node) {
	g.WriteString(n.Name)
}

func (g *Generator) StringNode(n *Node) {
	g.WriteString(strconv.Quote(n.Value.String()))
}
func (g *Generator) IntNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Int()))
}
func (g *Generator) UIntNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Uint()))
}
func (g *Generator) FloatNode(n *Node) {
	g.WriteString(fmt.Sprintf("%f", n.Value.Float()))
}
func (g *Generator) BoolNode(n *Node) {
	g.WriteString(fmt.Sprintf("%t", n.Value.Bool()))
}
func (g *Generator) PointerNode(n *Node) {
	g.WriteByte('&')
	g.WriteCode(n.nodes[0])
}

func (g *Generator) MapNode(n *Node) {
	g.WriteString(n.Name)
	g.WriteByte('{')
	for _, node := range n.nodes {
		g.WriteCode(node)
		g.WriteByte(':')
		g.WriteCode(node.nodes[0])
		g.WriteByte(',')
	}
	g.WriteByte('}')
}
func (g *Generator) SliceNode(n *Node) {
	g.WriteString(n.Name)
	g.WriteByte('{')
	for _, node := range n.nodes {
		g.WriteCode(node.nodes[0])
		g.WriteByte(',')
	}
	g.WriteByte('}')
}
func (g *Generator) StructNode(n *Node) {
	g.WriteString(n.Name)
	g.WriteByte('{')
	for _, node := range n.nodes {
		g.WriteString(node.Name)
		g.WriteByte(':')
		g.WriteCode(node.nodes[0])
		g.WriteByte(',')
	}
	g.WriteByte('}')
}
