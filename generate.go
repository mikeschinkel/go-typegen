package typegen

import (
	"fmt"
	"regexp"
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
	varnameCtr  int
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

func (g *Generator) NodeVarname(n *Node) string {
	if n.Varname() != "" {
		goto end
	}
	if n.NodeRef != nil {
		n.SetVarname(g.NodeVarname(n.NodeRef))
		goto end
	}
	g.varnameCtr++
	n.SetVarname(fmt.Sprintf("var%d", g.varnameCtr))
end:
	return n.Varname()
}

var pkgStripRE *regexp.Regexp

func (g *Generator) maybeStripPackage(name string) string {
	if name == "&" {
		goto end
	}
	if len(name) == 0 {
		goto end
	}
	if !strings.Contains(name, ".") {
		goto end
	}
	if pkgStripRE == nil {
		pkgStripRE = regexp.MustCompile(fmt.Sprintf(`^(\W*)%s\.`, g.omitPkg))
	}
	name = pkgStripRE.ReplaceAllString(name, "$1")
end:
	return name
}

func (g *Generator) WriteCode(n *Node) {
	n.Name = g.maybeStripPackage(n.Name)
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

	case IntNode:
		g.IntNode(n)
	case Int8Node:
		g.Int8Node(n)
	case Int16Node:
		g.Int16Node(n)
	case Int32Node:
		g.Int32Node(n)
	case Int64Node:
		g.Int64Node(n)
	case UintNode:
		g.UintNode(n)
	case Uint8Node:
		g.Uint8Node(n)
	case Uint16Node:
		g.Uint16Node(n)
	case Uint32Node:
		g.Uint32Node(n)
	case Uint64Node:
		g.Uint64Node(n)
	case Float32Node:
		g.Float32Node(n)
	case Float64Node:
		g.Float64Node(n)

	}
}

func (g *Generator) Int8Node(n *Node) {
	g.WriteString(fmt.Sprintf("int8(%d)", n.Value.Int()))
}

func (g *Generator) Int16Node(n *Node) {
	g.WriteString(fmt.Sprintf("int16(%d)", n.Value.Int()))
}

func (g *Generator) Int32Node(n *Node) {
	g.WriteString(fmt.Sprintf("int32(%d)", n.Value.Int()))
}

func (g *Generator) Int64Node(n *Node) {
	g.WriteString(fmt.Sprintf("int64(%d)", n.Value.Int()))
}

func (g *Generator) Uint8Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint8(%d)", n.Value.Uint()))
}

func (g *Generator) Uint16Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint16(%d)", n.Value.Uint()))
}

func (g *Generator) Uint32Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint32(%d)", n.Value.Uint()))
}

func (g *Generator) Uint64Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint64(%d)", n.Value.Uint()))
}

func (g *Generator) Float32Node(n *Node) {
	g.WriteString(fmt.Sprintf("float32(%f)", n.Value.Float()))
}

func (g *Generator) Float64Node(n *Node) {
	g.WriteString(fmt.Sprintf("float64(%f)", n.Value.Float()))
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

func (g *Generator) WriteAssignment(a *Assignment) {
	g.WriteString(fmt.Sprintf("%s%s %s %s\n",
		g.Indent,
		a.LHS,
		a.Op,
		a.RHS,
	))
}

func (g *Generator) recordAssignment(n *Node) {
	var parent *Node
	if n == nil {
		// We are basically at the root
		goto end
	}
	parent = n.parent
	if parent == nil {
		panic("Handle when node.Parent is nil")
	}
	switch parent.Type {
	case PointerNode:
		g.recordAssignment(parent.parent)
	case FieldNameNode, ElementNode:
		g.Assignments = append(g.Assignments, &Assignment{
			LHS: g.LHS(parent, n),
			Op:  g.Op(parent, n),
			RHS: g.RHS(parent, n),
		})
	default:
		panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
	}
end:
}

func (g *Generator) RefNode(n *Node) {
	node, isPtr, index := g.deref(n)
	if !g.varMap.HasVar(index) {
		// Var<Index> already been output, so we need to come back later and connect
		// property with variable.
		g.WriteString("nil")
		g.recordAssignment(n)
		goto end
	}
	// Write variable
	if isPtr {
		g.WriteByte('&')
	}
	g.WriteString(g.NodeVarname(node))
end:
}

func (g *Generator) StringNode(n *Node) {
	g.WriteString(strconv.Quote(n.Value.String()))
}
func (g *Generator) IntNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Int()))
}
func (g *Generator) UintNode(n *Node) {
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

func (g *Generator) Op(parent, node *Node) (op string) {
	switch parent.Type {
	case FieldNameNode, ElementNode:
		op = "="
	default:
		op = ":="
	}
	return op
}

func (g *Generator) LHS(parent, node *Node) (lhs string) {
	switch node.Type {
	case RefNode:
		lhs = g.LHS(parent, node.NodeRef)
	default:
		lhs = fmt.Sprintf("%s.%s", g.NodeVarname(parent.parent), parent.Name)
	}
	return lhs
}

func (g *Generator) RHS(parent, node *Node) (rhs string) {
	switch node.Type {
	case RefNode:
		rhs = g.RHS(parent, node.NodeRef)
	default:
		rhs = "&" + g.NodeVarname(node)
	}
	return rhs
}

func (g *Generator) deref(n *Node) (_ *Node, isPtr bool, index int) {
	var nr *Node
	if n.NodeRef == nil {
		goto end
	}
	nr = n.NodeRef
	if nr.Type == PointerNode && nr.NodeRef.parent == nil {
		nr.NodeRef.parent = nr
	}
	n, isPtr, index = g.deref(nr)
	if n.parent != nil && n.parent.Type == PointerNode {
		if n.parent.varname == "" {
			n.parent.varname = n.varname
		}
		index = n.Index
		n = n.parent
		isPtr = true
	}
end:
	return n, isPtr, index
}
