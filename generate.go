package typegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Generator struct {

	// Builder caches the output during generation, and is embedded so Generator
	// objects can call its methods directly.
	strings.Builder

	// Indent typically contains two spaces which are used to output code for
	// indentation, but can be replaced with (a) tab(s) or a different number of
	// spaces when this pacakge is used. NOTE: The tests assume two spaces.
	Indent string

	// omitPkg is the package name to be stripped from all types during code
	// generation. Since Go does not allow using the name of the current package as a
	// prefix, omitPkg allows code to be generated that does not include the current
	// package name where you plan to use the generated code (which I expect will be
	// a test package.)
	omitPkg string

	// genMap is a map that contains the nodes that have been generated so we can
	// avoid generating multiple times. It is keys by value of `Node.Index` and
	// its value will be the corresponding `*Node`.
	genMap GenMap

	// assignments is a slice of the assignments that need to be generated after for
	// each node from `CodeBuilder.nodeMap` is generated that needs to be generated.
	// These are registered in `Generator.registerAssignment()` from within
	// `Generator.RefNode()`, and then they are generated in `CodeBuilder.Generate()`
	// which calls `Generator.writeAssigment()`.
	assignments Assignments

	// varnameCtr keeps track of the next variable name suffix, e.g. `var`, `var2`,
	// `var3`, ... `varN``.  This is used in `Generator.nodeVarname()`
	varnameCtr int

	// prefixLen is set in `CodeBuilder.Generate()` to specify have make bytes it has
	// written to the embedded `strings.Builder` of this `Generator` so that
	// `Generator.RefNode()` can tell if the Generator has written any data or not.
	// If it has not then it should call `Generator.WriteCode()` on the current node
	// since it will be the first node, otherwise it would output `nil` for use in a
	// container property, are as a variable to be references if the code was already
	// generated.
	prefixLen int
}

// NewGenerator instantiates a new *Generator object with one param; the package
// name to omit from code generation which should be the package the code will be
// used within, e.g. "typegen_test" if the output is to be used for tests for the
// `typegen` package.
func NewGenerator(omitPkg string) *Generator {
	return &Generator{
		Indent:      "  ",
		omitPkg:     omitPkg,
		genMap:      make(GenMap),
		assignments: make(Assignments, 0),
	}
}

// WriteCode accepts a *Node and writes code to the embedded strings.Builder that
// will create that node. Note that it should only output one level and expect
// properties that are containers — array, slice, struct, ptr, map, etc. — to be
// generated separately. This function will add an `*Assignment` for each of
// those properties.
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

// RefNode either 1. Generates code for it's .NodeRef if no code generated yet,
// 2. uses the embedded `strings.Builder` to generate a `nil`, and then records
// an assignment for the node to be generated after this line is generated, o 3.
// uses the embedded `strings.Builder` to generate a varname with potential
// address of operator. RefNodes were designed to allow complex object graphs to
// be generated one container at a time by replacing expression code with a `nil`
// and then delegating the expression to a later variable assignment.
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
		// have missed.) So just assign a nil and register that we need to generate an
		// assignment of a pointer to the variable containing the value later.
		g.WriteString("nil")
		g.registerAssignment(n)
		goto end

	default:
		if n.NodeRef.Type == PointerNode {
			g.WriteByte('&')
		}
		g.WriteString(g.nodeVarname(n))

	}
end:
}

// Int8Node generates the int8 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Int8Node(n *Node) {
	g.WriteString(fmt.Sprintf("int8(%d)", n.Value.Int()))
}

// Int16Node generates the int16 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Int16Node(n *Node) {
	g.WriteString(fmt.Sprintf("int16(%d)", n.Value.Int()))
}

// Int32Node generates the int32 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Int32Node(n *Node) {
	g.WriteString(fmt.Sprintf("int32(%d)", n.Value.Int()))
}

// Int64Node generates the int64 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Int64Node(n *Node) {
	g.WriteString(fmt.Sprintf("int64(%d)", n.Value.Int()))
}

// Uint8Node generates the uint8 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Uint8Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint8(%d)", n.Value.Uint()))
}

// Uint16Node generates the uint16 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Uint16Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint16(%d)", n.Value.Uint()))
}

// Uint32Node generates the uint32 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Uint32Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint32(%d)", n.Value.Uint()))
}

// Uint64Node generates the uint64 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Uint64Node(n *Node) {
	g.WriteString(fmt.Sprintf("Uint64(%d)", n.Value.Uint()))
}

// Float32Node generates the float32 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Float32Node(n *Node) {
	g.WriteString(fmt.Sprintf("float32(%f)", n.Value.Float()))
}

// Float64Node generates the float64 code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) Float64Node(n *Node) {
	g.WriteString(fmt.Sprintf("float64(%f)", n.Value.Float()))
}

// StringNode generates the string code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) StringNode(n *Node) {
	g.WriteString(strconv.Quote(n.Value.String()))
}

// IntNode generates the Int code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) IntNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Int()))
}

// UintNode generates the Uint code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) UintNode(n *Node) {
	g.WriteString(fmt.Sprintf("%d", n.Value.Uint()))
}

// BoolNode generates the bool code from a Node using the embedded
// `strings.Builder.`
func (g *Generator) BoolNode(n *Node) {
	g.WriteString(fmt.Sprintf("%t", n.Value.Bool()))
}

// InterfaceNode generates the `any` code from a Node using the embedded
// `strings.Builder.` This generator generates `any` for data created as type
// `interface{}` since there is no way in Go to differentiate that I am aware of,
// and even if there is I don't think differentiating would be worth the effort
// when the use-case of this project is considered.
func (g *Generator) InterfaceNode(n *Node) {
	g.WriteString("any(")
	g.WriteCode(n.nodes[0])
	g.WriteByte('}')
}

// PointerNode generates the pointer code from a Node using the embedded
// `strings.Builder` ASSUMING it ever gets called.
//
//goland:noinspection GoUnusedParameter
func (g *Generator) PointerNode(*Node) {
	//if g.VarGenerated(n.nodes[0]) {
	//	// If var is already generated then g.RefNode() will output a variable name which
	//	// we'll need to get the address of with `&. But if not, it will output a `nil`
	//	// which we can't use `&` in front of.
	//	g.WriteByte('&')
	//}
	//g.WriteCode(n.nodes[0])
	panic("Verify this gets called")
}

// StructNode generates the struct code from a Node using the embedded
// `strings.Builder.`
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

// MapNode generates the map code from a Node using the embedded `strings.Builder.`
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

// ArrayNode generates the array code from a Node using the embedded `strings.Builder.`
func (g *Generator) ArrayNode(n *Node) {
	g.nodeElements(n)
}

// SliceNode generates the slice code from a Node using the embedded `strings.Builder.`
func (g *Generator) SliceNode(n *Node) {
	g.nodeElements(n)
}

// nodeElements generates the element's code for both arrays and slices using the
// embedded `strings.Builder.`
func (g *Generator) nodeElements(n *Node) {
	g.WriteString(n.Name)
	g.WriteByte('{')
	for _, node := range n.nodes {
		g.WriteCode(node.nodes[0])
		g.WriteByte(',')
	}
	g.WriteByte('}')
}

// InvalidNode generates the `nil` for invalid Nodes using the embedded
// `strings.Builder.` Taking a `reflect.ValueOf(nil)` will return an invalid
// reflect type so this is appropriate, although edge cases may reveal a need to
// handle them differently.
//
//goland:noinspection GoUnusedParameter
func (g *Generator) InvalidNode(*Node) {
	g.WriteString("nil")
}

// VarGenerated will return true if that variable for n.Index has already been
// generated and thus can be referenced by name.
//
//goland:noinspection GoUnusedParameter
func (g *Generator) VarGenerated(*Node) (pointing bool) {
	panic("Verify this is ever called")
	//	if n.Type != RefNode {
	//		goto end
	//	}
	//	pointing = g.wasGenerated(n)
	//end:
	return pointing
}

// ancestorVarname looks for the varname from the Node's parent, or its parent,
// or its parent, and so on recursively, until there is no more parents left,
// e.g. we get to the root of the data structure.
func (g *Generator) ancestorVarname(n *Node) (s string) {
	if n.parent == nil {
		goto end
	}
	if n.parent.varname != "" {
		s = n.parent.varname
		goto end
	}
	s = g.ancestorVarname(n.parent)
end:
	return s
}

// nodeVarname returns AND SETS the varname for the Node. NOTE that for Pointers
// and NodeRefs it dereferences first by calling itself recursively. Basically
// this reserves the variable name by incrementing a counter and when it sets the
// name it is basically reserving this name for this node, e.g. `var1`, var2`,
// etc. In future iterations we may allow developer-defined names, but only if
// this project gets a LOT of interest, which I kinda doubt will happen, or
// someone submits a PR, or someone pays me to do it. #fwiw
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

// lhs return the left-hand side for an assignment as a string, given a *Node
// This will always be the var name of the parent node, e.g. `var1` plus the
// property name of the struct to be assigned (or maybe not, we'll see if this
// assumption is wrong after we do
// // testing for more use-cases.
func (g *Generator) lhs(node *Node) (lhs string) {
	return fmt.Sprintf("%s.%s", g.ancestorVarname(node), node.parent.Name)
}

// assignOp will return assigment operator; an `=` if a field
func (g *Generator) assignOp(node *Node) (op string) {
	switch node.parent.Type {
	case FieldNode, ElementNode:
		op = "="
	default:
		op = ":="
	}
	return op
}

// rhs return the right-hand side for an assignment as a string, given a *Node.
// This will always be a pointer variable reference given the nature of the
// output (or maybe not, we'll see if this assumption is wrong after we do
// testing for more use-cases.
func (g *Generator) rhs(node *Node) (rhs string) {
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

// writeAssignment will write an assigment previously registered by
// `registerAssignment()`, which will be called in `CodeBuilder.Generate()` after
// the *Node in which it was register for is written.
func (g *Generator) writeAssignment(a *Assignment) {
	g.WriteString(fmt.Sprintf("%s%s %s %s\n",
		g.Indent,
		a.LHS,
		a.Op,
		a.RHS,
	))
}

// registerAssignment will take a node and register an assigment line to be
// generated after the current Node is being generated in
// `CodeBuilder.Generate()`. Assignment lines take on the form of `<LHS> <Op>
// <RHS>` e.g. `var1.prop = 10` or `var2 := []string{}`
func (g *Generator) registerAssignment(n *Node) {
	var parent *Node
	if n == nil {
		// We are at the root
		goto end
	}
	parent = n.parent
	if parent == nil {
		panic("Handle when node.Parent is nil")
	}
	switch parent.Type {
	case PointerNode:
		g.registerAssignment(parent.parent)
	case FieldNode, ElementNode:
		g.assignments = append(g.assignments, &Assignment{
			LHS: g.lhs(n),
			Op:  g.assignOp(n),
			RHS: g.rhs(n),
		})
	default:
		panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
	}
end:
}

// pkgStripRE is a regular expression used by Generator.maybeStripPackage()
var pkgStripRE *regexp.Regexp

// maybeStripPackage will remove `foo.` from `foo.Bar`, *foo.Bar`, []foo.Bar` and so on.
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
