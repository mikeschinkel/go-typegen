package typegen

import (
	"fmt"
	"strconv"
	"strings"
)

type CodeBuilder struct {

	// Builder caches the output during generation, and is embedded so CodeBuilder
	// objects can call its methods directly.
	strings.Builder

	// Indent typically contains two spaces which are used to output code for
	// indentation, but can be replaced with (a) tab(s) or a different number of
	// spaces when this package is used. NOTE: The tests assume two spaces.
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
	// each node from `NodeMarshaler.nodeMap` is generated that needs to be generated.
	// These are registered in `CodeBuilder.registerAssignment()` from within
	// `CodeBuilder.RefNode()`, and then they are generated in `NodeMarshaler.Build()`
	// which calls `CodeBuilder.writeAssigment()`.
	assignments Assignments

	// varnameCtr keeps track of the next variable name suffix, e.g. `var`, `var2`,
	// `var3`, ... `varN``.  This is used in `CodeBuilder.nodeVarname()`
	varnameCtr int

	// prefixLen is set in `NodeMarshaler.Build()` to specify have make bytes it has
	// written to the embedded `strings.Builder` of this `CodeBuilder` so that
	// `CodeBuilder.RefNode()` can tell if the CodeBuilder has written any data or not.
	// If it has not then it should call `CodeBuilder.WriteCode()` on the current node
	// since it will be the first node, otherwise it would output `nil` for use in a
	// container property, are as a variable to be references if the code was already
	// generated.
	prefixLen int

	nodes    Nodes
	funcName string
}

// NewCodeBuilder instantiates a new *CodeBuilder object with one param; the package
// name to omit from code generation which should be the package the code will be
// used within, e.g. "typegen_test" if the output is to be used for tests for the
// `typegen` package.
func NewCodeBuilder(funcName, omitPkg string, nodes Nodes) *CodeBuilder {
	return &CodeBuilder{
		Indent:      "  ",
		omitPkg:     omitPkg,
		funcName:    funcName,
		nodes:       nodes,
		genMap:      make(GenMap),
		assignments: make(Assignments, 0),
	}
}

func (b *CodeBuilder) NodeCount() int {
	return len(b.nodes) - 1
}

// selectNode selects a node for use in Generate(). If it is a pointer it
// dereferences it by selecting the next node in the list of .nodes and
// increments and returns the index. It also returns if it was a pointer
// so that Generate() can generate a pointer return value, if so.
func (b *CodeBuilder) selectNode(index int) (n *Node, _ int, nt NodeType) {
	n = b.nodes[index]
	if n.Type != PointerNode {
		goto end
	}
	nt = n.Type
	if index >= b.NodeCount() {
		// We are at the element node, but it is a pointer or interface.
		n = n.nodes[0]
		goto end
	}
	index++
	n = b.nodes[index]
end:
	return n, index, nt
}

func (b *CodeBuilder) String() string {
	return b.Build()
}

func (b *CodeBuilder) Build() string {
	var returnVar, returnType string
	var n *Node
	var nt NodeType

	nodeCnt := b.NodeCount()
	for i := 1; i <= nodeCnt; i++ {
		n, i, nt = b.selectNode(i)
		if b.wasGenerated(n) {
			// n is pointed at by prior, so we've already output it
			continue
		}
		if returnVar == "" {
			returnVar, returnType = b.returnVarAndType(n, nt)
		}
		b.WriteString(fmt.Sprintf("%s%s := ", b.Indent, b.nodeVarname(n)))
		b.prefixLen = b.Builder.Len()
		b.WriteCode(n)
		b.WriteByte('\n')

		// Record that this var has been generated
		b.genMap[n.Index] = n

	}
	for _, a := range b.assignments {
		b.writeAssignment(a)
	}
	b.WriteString(fmt.Sprintf("%sreturn %s\n", b.Indent, returnVar))
	b.WriteByte('}')
	return fmt.Sprintf("func %s() %s {\n%s",
		b.funcName,
		returnType,
		b.Builder.String(),
	)
}

// WriteCode accepts a *Node and writes code to the embedded strings.Builder that
// will create that node. Note that it should only output one level and expect
// properties that are containers — array, slice, struct, ptr, map, etc. — to be
// generated separately. This function will add an `*Assignment` for each of
// those properties.
func (b *CodeBuilder) WriteCode(n *Node) {
	n.Name = maybeStripPackage(n.Name, b.omitPkg)
	n.Name = replaceInterfaceWithAny(n.Name)
	n.resetDebugString()
	switch n.Type {
	case RefNode:
		b.RefNode(n)
	case StructNode:
		b.StructNode(n)
	case PointerNode:
		b.PointerNode(n)
	case MapNode:
		b.MapNode(n)
	case ArrayNode:
		b.ArrayNode(n)
	case InterfaceNode:
		b.InterfaceNode(n)
	case StringNode:
		b.StringNode(n)
	case BoolNode:
		b.BoolNode(n)
	case InvalidNode:
		b.InvalidNode(n)
	case SliceNode:
		b.SliceNode(n)

	case IntNode:
		b.IntNode(n)
	case Int8Node:
		b.Int8Node(n)
	case Int16Node:
		b.Int16Node(n)
	case Int32Node:
		b.Int32Node(n)
	case Int64Node:
		b.Int64Node(n)
	case UintNode:
		b.UintNode(n)
	case Uint8Node:
		b.Uint8Node(n)
	case Uint16Node:
		b.Uint16Node(n)
	case Uint32Node:
		b.Uint32Node(n)
	case Uint64Node:
		b.Uint64Node(n)
	case Float32Node:
		b.Float32Node(n)
	case Float64Node:
		b.Float64Node(n)
	case UnsafePointerNode:
		b.UnsafePointerNode(n)

	}
}

// RefNode either 1. Generates code for it's .NodeRef if no code generated yet,
// 2. uses the embedded `strings.Builder` to generate a `nil`, and then records
// an assignment for the node to be generated after this line is generated, o 3.
// uses the embedded `strings.Builder` to generate a varname with potential
// address of operator. RefNodes were designed to allow complex object graphs to
// be generated one container at a time by replacing expression code with a `nil`
// and then delegating the expression to a later variable assignment.
func (b *CodeBuilder) RefNode(n *Node) {
	switch {
	case b.Builder.Len() == b.prefixLen:
		// Output has not been generated for any node so this is the first node and the
		// node to which this node is a reference must have its code generated. Also,
		// this path should only be taken once because if not we'll be in an infinite
		// recursion. This can happen when a container contains a value that contains a
		// pointer back to the original container.
		b.WriteCode(n.NodeRef)

	case !b.wasGenerated(n):
		// Output has not been generated for this node which means it is being assigned
		// to a property of a struct, or as an element of a map, slice or array (I think
		// that is exhaustive of when this should run but there may be some other cases I
		// have missed.) So just assign a nil and register that we need to generate an
		// assignment of a pointer to the variable containing the value later.
		b.WriteString("nil")
		b.registerAssignment(n)
		goto end

	default:
		if n.NodeRef.Type == PointerNode {
			b.WriteByte('&')
		}
		b.WriteString(b.nodeVarname(n))

	}
end:
}

// Int8Node generates the int8 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Int8Node(n *Node) {
	b.WriteString(fmt.Sprintf("int8(%d)", n.Value.Int()))
}

// Int16Node generates the int16 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Int16Node(n *Node) {
	b.WriteString(fmt.Sprintf("int16(%d)", n.Value.Int()))
}

// Int32Node generates the int32 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Int32Node(n *Node) {
	b.WriteString(fmt.Sprintf("int32(%d)", n.Value.Int()))
}

// Int64Node generates the int64 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Int64Node(n *Node) {
	b.WriteString(fmt.Sprintf("int64(%d)", n.Value.Int()))
}

// Uint8Node generates the uint8 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Uint8Node(n *Node) {
	b.WriteString(fmt.Sprintf("Uint8(%d)", n.Value.Uint()))
}

// Uint16Node generates the uint16 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Uint16Node(n *Node) {
	b.WriteString(fmt.Sprintf("Uint16(%d)", n.Value.Uint()))
}

// Uint32Node generates the uint32 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Uint32Node(n *Node) {
	b.WriteString(fmt.Sprintf("Uint32(%d)", n.Value.Uint()))
}

// Uint64Node generates the uint64 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Uint64Node(n *Node) {
	b.WriteString(fmt.Sprintf("Uint64(%d)", n.Value.Uint()))
}

// Float32Node generates the float32 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Float32Node(n *Node) {
	b.WriteString(fmt.Sprintf("float32(%f)", n.Value.Float()))
}

// Float64Node generates the float64 code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) Float64Node(n *Node) {
	b.WriteString(fmt.Sprintf("float64(%f)", n.Value.Float()))
}

// StringNode generates the string code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) StringNode(n *Node) {
	b.WriteString(strconv.Quote(n.Value.String()))
}

// IntNode generates the Int code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) IntNode(n *Node) {
	b.WriteString(fmt.Sprintf("%d", n.Value.Int()))
}

// UintNode generates the Uint code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) UintNode(n *Node) {
	b.WriteString(fmt.Sprintf("%d", n.Value.Uint()))
}

// BoolNode generates the bool code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) BoolNode(n *Node) {
	b.WriteString(fmt.Sprintf("%t", n.Value.Bool()))
}

func (b *CodeBuilder) UnsafePointerNode(n *Node) {
	//b.WriteString(fmt.Sprintf("%d", n.Value.UnsafePointer()))
	// Should not output a real unsafePointer
	// TODO: Find a way to handle this better
	b.WriteString("-1")
}

// InterfaceNode generates the `any` code from a Node using the embedded
// `strings.Builder.` This generator generates `any` for data created as type
// `interface{}` since there is no way in Go to differentiate that I am aware of,
// and even if there is I don't think differentiating would be worth the effort
// when the use-case of this project is considered.
func (b *CodeBuilder) InterfaceNode(n *Node) {
	b.WriteString("any(")
	b.WriteCode(n.nodes[0])
	b.WriteByte(')')
}

// PointerNode generates the pointer code from a Node using the embedded
// `strings.Builder` ASSUMING it ever gets called.
//
//goland:noinspection GoUnusedParameter
func (b *CodeBuilder) PointerNode(*Node) {
	//if b.varGenerated(n.nodes[0]) {
	//	// If var is already generated then b.RefNode() will output a variable name which
	//	// we'll need to get the address of with `&. But if not, it will output a `nil`
	//	// which we can't use `&` in front of.
	//	b.WriteByte('&')
	//}
	//b.WriteCode(n.nodes[0])
	panic("Verify this gets called")
}

// StructNode generates the struct code from a Node using the embedded
// `strings.Builder.`
func (b *CodeBuilder) StructNode(n *Node) {
	b.WriteString(n.Name)
	b.WriteByte('{')
	for _, node := range n.nodes {
		b.WriteString(node.Name)
		b.WriteByte(':')
		b.WriteCode(node.nodes[0])
		b.WriteByte(',')
	}
	b.WriteByte('}')
}

// MapNode generates the map code from a Node using the embedded `strings.Builder.`
func (b *CodeBuilder) MapNode(n *Node) {
	b.WriteString(n.Name)
	b.WriteByte('{')
	for _, node := range n.nodes {
		b.WriteCode(node)
		b.WriteByte(':')
		b.WriteCode(node.nodes[0])
		b.WriteByte(',')
	}
	b.WriteByte('}')
}

// ArrayNode generates the array code from a Node using the embedded `strings.Builder.`
func (b *CodeBuilder) ArrayNode(n *Node) {
	b.nodeElements(n)
}

// SliceNode generates the slice code from a Node using the embedded `strings.Builder.`
func (b *CodeBuilder) SliceNode(n *Node) {
	b.nodeElements(n)
}

// nodeElements generates the element's code for both arrays and slices using the
// embedded `strings.Builder.`
func (b *CodeBuilder) nodeElements(n *Node) {
	b.WriteString(n.Name)
	b.WriteByte('{')
	for _, node := range n.nodes {
		b.WriteCode(node.nodes[0])
		b.WriteByte(',')
	}
	b.WriteByte('}')
}

// InvalidNode generates the `nil` for invalid Nodes using the embedded
// `strings.Builder.` Taking a `reflect.ValueOf(nil)` will return an invalid
// reflect type so this is appropriate, although edge cases may reveal a need to
// handle them differently.
//
//goland:noinspection GoUnusedParameter
func (b *CodeBuilder) InvalidNode(*Node) {
	b.WriteString("nil")
}

// varGenerated will return true if that variable for n.Index has already been
// generated and thus can be referenced by name.
//
//goland:noinspection GoUnusedParameter
func (b *CodeBuilder) varGenerated(*Node) (pointing bool) {
	panic("Verify this is ever called")
	//	if n.Type != RefNode {
	//		goto end
	//	}
	//	pointing = b.wasGenerated(n)
	//end:
	return pointing
}

// ancestorVarname looks for the varname from the Node's parent, or its parent,
// or its parent, and so on recursively, until there is no more parents left,
// e.g. we get to the root of the data structure.
func (b *CodeBuilder) ancestorVarname(n *Node) (s string) {
	if n.parent == nil {
		goto end
	}
	if n.parent.varname != "" {
		s = n.parent.varname
		goto end
	}
	s = b.ancestorVarname(n.parent)
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
func (b *CodeBuilder) nodeVarname(n *Node) string {
	if n.varname != "" {
		goto end
	}
	if n.Type == PointerNode {
		n.SetVarname(b.nodeVarname(n.nodes[0]))
		goto end
	}
	if n.NodeRef != nil {
		n.SetVarname(b.nodeVarname(n.NodeRef))
		goto end
	}
	b.varnameCtr++
	n.SetVarname(fmt.Sprintf("var%d", b.varnameCtr))
end:
	return n.varname
}

// lhs return the left-hand side for an assignment as a string, given a *Node
// This will always be the var name of the parent node, e.g. `var1` plus the
// property name of the struct to be assigned (or maybe not, we'll see if this
// assumption is wrong after we do
// // testing for more use-cases.
func (b *CodeBuilder) lhs(node *Node) (lhs string) {
	if node.parent != nil && node.parent.Type == ElementNode {
		// Handles `var1[0]` of []any{1, 2, 3} for var1[0] = var2
		lhs = fmt.Sprintf("%s[%d]", b.ancestorVarname(node), node.Index)
		goto end
	}
	lhs = fmt.Sprintf("%s.%s", b.ancestorVarname(node), node.parent.Name)
end:
	return lhs
}

// assignOp will return assigment operator; an `=` if a field
func (b *CodeBuilder) assignOp(node *Node) (op string) {
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
func (b *CodeBuilder) rhs(node *Node) (rhs string) {
	if b.omitAddressOf(node) {
		return b.nodeVarname(node)
	}
	return "&" + b.nodeVarname(node)
}

// omitAddressOf returns true if we should omit the address of operator (&) for
// the right-hand side.
func (b *CodeBuilder) omitAddressOf(node *Node) (omit bool) {

	if node.NodeRef == nil {
		goto end
	}
	if len(node.NodeRef.nodes) == 0 {
		goto end
	}
	if node.NodeRef.Type == PointerNode {
		goto end
	}
	if node.NodeRef.nodes[0].Type == PointerNode {
		goto end
	}
	omit = true
end:
	return omit
}

// wasGenerated returns true if the node has already been generated
func (b *CodeBuilder) wasGenerated(node *Node) (generated bool) {

	if len(b.genMap) == 0 {
		goto end
	}

	_, generated = b.genMap[node.Index]
	if generated {
		goto end
	}

	if b.findPointedToNode(node) {
		generated = true
		goto end
	}
end:
	return generated
}

// findPointedToNode returns true when the node passed is found with a parent
// whose type is a pointer. It dereferences both Ref and Pointer nodes before
// looking for a match.
func (b *CodeBuilder) findPointedToNode(n *Node) (found bool) {
	switch {
	case n.Type == RefNode && n.NodeRef != nil:
		found = b.findPointedToNode(n.NodeRef)
		goto end

	case n.Type == PointerNode && len(n.nodes) != 0:
		found = b.findPointedToNode(n.nodes[0])
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
func (b *CodeBuilder) returnVarAndType(n *Node, nt NodeType) (rv, rt string) {
	switch nt {
	case PointerNode:
		rv += "&" + b.nodeVarname(n)
		rt = "*" + maybeStripPackage(n.Value.Type().String(), b.omitPkg)
		goto end
	case InterfaceNode:
		fallthrough
	default:
		rv = b.nodeVarname(n)
		rt = "error" // error is a built-in type that can can be nil.
		if n.Value.IsValid() {
			// Get the return type, and with `.omitPkg` package stripped, if applicable
			rt = maybeStripPackage(n.Value.Type().String(), b.omitPkg)
			rt = replaceInterfaceWithAny(rt)
		}
	}
end:
	return rv, rt
}

// writeAssignment will write an assigment previously registered by
// `registerAssignment()`, which will be called in `NodeMarshaler.Build()` after
// the *Node in which it was register for is written.
func (b *CodeBuilder) writeAssignment(a *Assignment) {
	b.WriteString(fmt.Sprintf("%s%s %s %s\n",
		b.Indent,
		a.LHS,
		a.Op,
		a.RHS,
	))
}

// registerAssignment will take a node and register an assigment line to be
// generated after the current Node is being generated in
// `NodeMarshaler.Build()`. Assignment lines take on the form of `<LHS> <Op>
// <RHS>` e.g. `var1.prop = 10` or `var2 := []string{}`
func (b *CodeBuilder) registerAssignment(n *Node) {
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
		b.registerAssignment(parent.parent)
	case FieldNode, ElementNode:
		b.assignments = append(b.assignments, &Assignment{
			LHS: b.lhs(n),
			Op:  b.assignOp(n),
			RHS: b.rhs(n),
		})
	default:
		panicf("Node type '%s' not implemented", nodeTypeName(parent.Type))
	}
end:
}
