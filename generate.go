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
	genMap      GenMap
	Assignments Assignments
	varnameCtr  int
	prefixLen   int
}

func NewGenerator(omitPkg string) *Generator {
	return &Generator{
		Indent:      "  ",
		omitPkg:     omitPkg,
		genMap:      make(GenMap),
		Assignments: make(Assignments, 0),
	}
}

func (g *Generator) WriteCode(n *Node) {
	n.Name = g.maybeStripPackage(n.Name)
	n.resetDebugString()
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

func (g *Generator) RefNode(n *Node) {
	switch {
	case g.Builder.Len() == g.prefixLen:
		// Output has not been generated for any node so this is the first node and the
		// node to which this node is a reference must have its code generated. Also,
		// this path should only be taken once because if not we'll be in an infinite
		// recursion. This can happen when a container contains a value that contains a
		// pointer back to the original container.
		g.WriteCode(n.NodeRef)

	case !g.wasGenerated(n):
		// Output has not been generated for this node which means it is being assigned
		// to a property of a struct, or as an element of a map, slice or array (I think
		// that is exhaustive of when this should run but there may be some other cases I
		// have missed.) So just assign a nil and record that we need to generate an
		// assignment of a pointer to the variable containing the value later.
		g.WriteString("nil")
		g.recordAssignment(n)
		goto end

	default:
		if n.NodeRef.Type == PointerNode {
			g.WriteByte('&')
		}
		g.WriteString(g.nodeVarname(n))

	}
end:
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

func (g *Generator) InterfaceNode(n *Node) {
	g.WriteString("any(")
	g.WriteCode(n.nodes[0])
	g.WriteByte('}')
}

func (g *Generator) InvalidNode(n *Node) {
	// TODO: Confirm this works in all cases.
	g.WriteString("nil")
}

func (g *Generator) StringNode(n *Node) {
	g.WriteString(strconv.Quote(n.Value.String()))
}
func (g *Generator) IntNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Int()))
}
func (g *Generator) UintNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Uint()))
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

func (g *Generator) ArrayNode(n *Node) {
	g.nodeElements(n)
}

func (g *Generator) SliceNode(n *Node) {
	g.nodeElements(n)
}

func (g *Generator) nodeElements(n *Node) {
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
	pointing = g.wasGenerated(n)
end:
	return pointing
}

func (g *Generator) parentVarname(n *Node) (s string) {
	if n.parent == nil {
		goto end
	}
	if n.parent.varname != "" {
		s = n.parent.varname
		goto end
	}
	s = g.parentVarname(n.parent)
end:
	return s
}

func (g *Generator) nodeVarname(n *Node) string {
	if n.varname != "" {
		goto end
	}
	if n.Type == PointerNode {
		n.SetVarname(g.nodeVarname(n.nodes[0]))
		goto end
	}
	if n.NodeRef != nil {
		n.SetVarname(g.nodeVarname(n.NodeRef))
		goto end
	}
	g.varnameCtr++
	n.SetVarname(fmt.Sprintf("var%d", g.varnameCtr))
end:
	return n.varname
}

//goland:noinspection GoUnusedParameter
func (g *Generator) assignOp(node *Node) (op string) {
	switch node.parent.Type {
	case FieldNode, ElementNode:
		op = "="
	default:
		op = ":="
	}
	return op
}

func (g *Generator) lhs(node *Node) (lhs string) {
	return fmt.Sprintf("%s.%s", g.parentVarname(node), node.parent.Name)
}

func (g *Generator) rhs(node *Node) (rhs string) {
	// This works for node.Type == RefNode, node.NodeRef.Type==PointerNode,
	// node.NodeRef.varname="varN", and node.parent.Type==FieldNode.
	// We'll need to handle others I am sure.
	return "&" + g.nodeVarname(node)
}

// wasGenerated returns true if the node has already been generated
func (g *Generator) wasGenerated(node *Node) (generated bool) {

	if len(g.genMap) == 0 {
		goto end
	}

	_, generated = g.genMap[node.Index]
	if generated {
		goto end
	}

	if g.findPointedToNode(node) {
		generated = true
		goto end
	}
end:
	return generated
}

// findPointedToNode returns true when the node passed is found with a parent
// whose type is a pointer. It dereferences both Ref and Pointer nodes before
// looking for a match.
func (g *Generator) findPointedToNode(n *Node) (found bool) {
	switch {
	case n.Type == RefNode && n.NodeRef != nil:
		found = g.findPointedToNode(n.NodeRef)
		goto end

	case n.Type == PointerNode && len(n.nodes) != 0:
		found = g.findPointedToNode(n.nodes[0])
		goto end

	case n.NodeRef == nil && len(n.nodes) == 0:
		goto end

	case n.parent != nil && n.parent.Type == PointerNode:
		found = true

	}
end:
	return found
}

// returnVarAndType will return the return variable and its type for the node received.
func (g *Generator) returnVarAndType(n *Node, isPtr bool) (rv, rt string) {
	if isPtr {
		rv += "&" + g.nodeVarname(n)
		rt = "*" + g.maybeStripPackage(n.Value.Type().String())
		goto end
	}
	rv = g.nodeVarname(n)
	rt = "error" // error is a built-in type that can can be nil.
	if n.Value.IsValid() {
		// Get the return type, and with `.omitPkg` package stripped, if applicable
		rt = g.maybeStripPackage(n.Value.Type().String())
	}
end:
	return rv, rt
}

func (g *Generator) writeAssignment(a *Assignment) {
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
	case FieldNode, ElementNode:
		g.Assignments = append(g.Assignments, &Assignment{
			LHS: g.lhs(n),
			Op:  g.assignOp(n),
			RHS: g.rhs(n),
		})
	default:
		panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
	}
end:
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
