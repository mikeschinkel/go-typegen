package typegen

import (
	"fmt"
	"strconv"
	"strings"
)

type Generator struct {
	strings.Builder
	Indent      string
	Depth       int
	omitPkg     string
	varMap      VarMap
	Assignments Assignments
}
type VarMap map[int]struct{}

func (m VarMap) HasVar(varNo int) bool {
	_, ok := m[varNo]
	return ok
}

func NewGenerator(omitPkg string) *Generator {
	return &Generator{
		Indent:      "  ",
		omitPkg:     omitPkg,
		varMap:      make(VarMap),
		Assignments: make(Assignments, 0),
	}
}

func (g *Generator) MaybeStripPackage(name string) string {
	if strings.HasPrefix(name, g.omitPkg+".") {
		name = name[len(g.omitPkg)+1:]
	}
	return name
}

func (g *Generator) WriteCode(n *Node) {
	n.Name = g.MaybeStripPackage(n.Name)
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
	case MapKeyNode:
		g.MapKeyNode(n)
	case MapValueNode:
		g.MapValueNode(n)

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

func (g *Generator) MapKeyNode(n *Node) {
	panic("Implement me")
}

func (g *Generator) MapValueNode(n *Node) {
	panic("Implement me")
}

type Assignments []*Assignment
type Assignment struct {
	LHS string
	RHS string
}

func (g *Generator) WriteAssignment(a *Assignment) {
	g.WriteString(fmt.Sprintf("%s%s = %s\n", g.Indent, a.LHS, a.RHS))
}

func (g *Generator) recordAssignment(n *Node) {
	parent := n.parent
	switch parent.Type {
	case FieldNameNode:
		varName := n.Varname()
		g.Assignments = append(g.Assignments, &Assignment{
			LHS: fmt.Sprintf("%s.%s", varName, parent.Name),
			RHS: "&" + varName,
		})
	case PointerNode:
		parent := parent.parent
		switch parent.Type {
		case FieldNameNode:
			varName := n.Varname()
			g.Assignments = append(g.Assignments, &Assignment{
				LHS: fmt.Sprintf("%s.%s", varName, parent.Name),
				RHS: "&" + varName,
			})
		default:
			panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
		}
	default:
		panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
	}
}

func (g *Generator) RefNode(n *Node) {
	if !g.varMap.HasVar(n.Index) {
		// Var<Index> already been output, so we need to come back later and connect
		// property with variable.
		g.WriteString("nil")
		g.recordAssignment(n)
		goto end
	}
	// Write variable
	g.WriteString(n.Name)
end:
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
	if g.VarGenerated(n.nodes[0]) {
		// If var is already generated then g.RefNode() will output a variable name which
		// we'll need to get the address of with `&. But if not, it will output a `nil`
		// which we can't use `&` in front of.
		g.WriteByte('&')
	}
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

// VarGenerated will return true if that variable for n.Index has already been
// generated and thus can be referenced by name.
func (g *Generator) VarGenerated(n *Node) (pointing bool) {
	if n.Type != RefNode {
		goto end
	}
	pointing = g.varMap.HasVar(n.Index)
end:
	return pointing
}
