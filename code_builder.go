package typegen

import (
	"fmt"
	"reflect"
	"sort"
)

type refVal = reflect.Value

type PointerMap map[uintptr]*Node

type CodeBuilder struct {
	funcName string
	value    reflect.Value
	original any
	nodeMap  NodeMap
	nodes    Nodes
	ptrMap   PointerMap
	root     *Node
	omitPkg  string
}

func NewCodeBuilder(value any, funcName string, omitPkg string) *CodeBuilder {
	cb := &CodeBuilder{
		original: value,
		value:    reflect.ValueOf(value),
		funcName: funcName,
		omitPkg:  omitPkg,
		nodeMap:  make(NodeMap),
		ptrMap:   make(PointerMap),
		nodes:    make(Nodes, 1), // Zero element is unused so node.index==0 can represent invalid
	}

	if cb.value.Kind() == reflect.Struct {
		panic("CodeBuilder currently does not support generating code for non-pointer structs. Pass a pointer to the struct instead.")
	}

	return cb
}

func (cb *CodeBuilder) Nodes() Nodes {
	return cb.nodes
}

func (cb *CodeBuilder) Build() {
	cb.root = cb.marshalValue(cb.value)
	if cb.NodeCount() == 0 {
		// If the root value is not a container, and thus not yet registered, register
		// the one value, so it can be converted in .String().
		cb.register(cb.value, cb.root)
	}
	cb.maybeReuniteNodes()
}

func (cb *CodeBuilder) NodeCount() int {
	return len(cb.nodeMap)
}

func (cb *CodeBuilder) maybeCollapseNodeRef(n *Node) {
	if n.Type != PointerNode {
		goto end
	}
	if n.nodes[0].Type != RefNode {
		goto end
	}
	// If it is a pointer and the first child is a RefNode, collapse the pointer's
	// NodeRef to point to a real node and not a RefNode.
	n.nodes[0] = n.nodes[0].NodeRef
end:
}

func (cb *CodeBuilder) String() string {
	return cb.Generate()
}

// selectNode selects a node for use in Generate(). If it is a pointer it
// dereferences it by selecting the next node in the list of .nodes and
// increments and returns the index. It also returns if it was a pointer
// so that Generate() can generate a pointer return value, if so.
func (cb *CodeBuilder) selectNode(index int) (n *Node, _ int, wasPtr bool) {
	n = cb.nodes[index]
	if n.Type != PointerNode {
		goto end
	}
	wasPtr = true
	if index >= cb.NodeCount() {
		goto end
	}
	index++
	n = cb.nodes[index]
end:
	return n, index, wasPtr
}

func (cb *CodeBuilder) Generate() string {
	var returnVar, returnType string
	var n *Node
	var wasPtr bool

	g := NewGenerator(cb.omitPkg)
	nodeCnt := cb.NodeCount()
	for i := 1; i <= nodeCnt; i++ {
		n, i, wasPtr = cb.selectNode(i)
		if g.wasGenerated(n) {
			// n is pointed at by prior, so we've already output it
			continue
		}
		if returnVar == "" {
			returnVar, returnType = g.returnVarAndType(n, wasPtr)
		}
		g.WriteString(fmt.Sprintf("%s%s := ", g.Indent, g.nodeVarname(n)))
		g.prefixLen = g.Builder.Len()
		g.WriteCode(n)
		g.WriteByte('\n')

		// Record that this var has been generated
		g.genMap[n.Index] = n

	}
	for _, a := range g.Assignments {
		g.writeAssignment(a)
	}
	g.WriteString(fmt.Sprintf("%sreturn %s\n", g.Indent, returnVar))
	g.WriteByte('}')
	return fmt.Sprintf("func %s() %s {\n%s", cb.funcName, returnType, g.String())
}

func (cb *CodeBuilder) marshalValue(rv refVal) (node *Node) {
	node = cb.marshalContainers(rv)
	if node != nil {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        "scalar",
		CodeBuilder: cb,
		Value:       rv,
	})
end:
	return node
}

func (cb *CodeBuilder) marshalContainers(rv refVal) (node *Node) {

	switch rv.Kind() {
	case reflect.Ptr:
		node = cb.marshalPtr(rv)
	case reflect.Struct:
		node = cb.marshalStruct(rv)
	case reflect.Slice:
		node = cb.marshalSlice(rv)
	case reflect.Map:
		node = cb.marshalMap(rv)
	case reflect.Interface:
		node = cb.marshalInterface(rv)
	case reflect.Array:
		node = cb.marshalArray(rv)
	default:
		goto end
	}
end:
	return node
}

// marshalArray marshals an array value to create a Node
func (cb *CodeBuilder) marshalArray(rv refVal) (node *Node) {
	return cb.marshalElements(rv, func() string {
		return fmt.Sprintf("[%d]%s", rv.Len(), rv.Type().Elem())
	})
}

// marshalSlice marshals a slice value to create a Node
func (cb *CodeBuilder) marshalSlice(rv refVal) (node *Node) {
	return cb.marshalElements(rv, func() string {
		return fmt.Sprintf("[]%s", rv.Type().Elem())
	})
}

// marshalElements marshals both array and slice values to create Nodes
func (cb *CodeBuilder) marshalElements(rv refVal, nameFunc func() string) (node *Node) {
	var ref refVal

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        nameFunc(),
		CodeBuilder: cb,
		Value:       rv,
	})
	ref = cb.register(rv, node)

	node.SetNodeCount(rv.Len())
	for i := 0; i < rv.Len(); i++ {
		child := NewNode(&NodeArgs{
			Name:        fmt.Sprintf("Index %d", i),
			Type:        ElementNode,
			CodeBuilder: cb,
			Value:       reflect.ValueOf(i),
			Index:       i,
		})
		node.AddNode(child)
		childValue := cb.marshalValue(rv.Index(i))
		childValue.Index = i
		childValue.Name = fmt.Sprintf("Value %d", i)
		childValue.resetDebugString()
		child.AddNode(childValue)
	}
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) marshalMap(rv refVal) (node *Node) {
	var name string
	var ref refVal

	var keys []reflect.Value

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	name = fmt.Sprintf("map[%s]%s", rv.Type().Key(), rv.Type().Elem())
	node = NewNode(&NodeArgs{
		Name:        name,
		CodeBuilder: cb,
		Value:       rv,
	})
	ref = cb.register(rv, node)
	keys = cb.sortedKeys(rv)
	node.SetNodeCount(len(keys))
	for _, key := range keys {
		child := cb.marshalValue(key)
		node.AddNode(child)
		child.AddNode(cb.marshalValue(rv.MapIndex(key)))
	}
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) marshalPtr(rv refVal) (node *Node) {
	var ref refVal

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:        "nil",
			CodeBuilder: cb,
			Value:       rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        "&",
		CodeBuilder: cb,
		Value:       rv,
	})
	ref = cb.register(rv, node)
	node.AddNode(cb.marshalValue(rv.Elem()))
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) marshalInterface(rv refVal) (node *Node) {
	var ref refVal

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:        "nil",
			CodeBuilder: cb,
			Value:       rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        rv.Type().String(),
		CodeBuilder: cb,
		Value:       rv,
	})
	ref = cb.register(rv, node)
	node.AddNode(cb.marshalValue(rv.Elem()))
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) marshalStruct(rv refVal) (node *Node) {
	var ref refVal

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        rv.Type().String(),
		CodeBuilder: cb,
		Value:       rv,
	})
	ref = cb.register(rv, node)
	for i := 0; i < rv.NumField(); i++ {
		name := rv.Type().Field(i).Name
		child := NewNode(&NodeArgs{
			Name:        name,
			Type:        FieldNameNode,
			CodeBuilder: cb,
			Value:       reflect.ValueOf(name),
			Index:       i,
		})
		node.AddNode(child)
		child.AddNode(cb.marshalValue(rv.Field(i)))
	}
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) findNodeMapKey(rv refVal) (node *Node, found bool) {
	var n *Node
	var k reflect.Value

	// First look for direct reflect.Value matches
	for k, n = range cb.nodeMap {
		if rv.Kind() == reflect.Pointer {
			continue
		}
		if !isSame(k, rv) {
			continue
		}
		if len(n.nodes) == 0 {
			continue
		}
		found = true
		goto end
	}

	// Next look for ptr>reflect.value matching reflect.value
	for k, n = range cb.nodeMap {
		if rv.Kind() != reflect.Pointer {
			continue
		}
		if !isSame(k, rv.Elem()) {
			continue
		}
		if len(n.nodes) == 0 {
			continue
		}
		n = n.nodes[0]
		found = true
		goto end
	}
end:
	if found {
		node = n
	}
	return node, found
}

func (cb *CodeBuilder) isRegistered(rv refVal) (node *Node, found bool) {

	if rv.Kind() != reflect.Pointer {
		node, found = cb.findNodeMapKey(rv)
		goto end
	}

	node, found = cb.ptrMap[rv.Pointer()]
	if found {
		// If the value of `rv` is a pointer, and we previously recorded it, then skip
		// registration.
		goto end
	}

	// Look for the value pointed to having already been registered
	node, found = cb.findNodeMapKey(rv)
	if found {
		// If yes, create a RefNode for it
		node = cb.newRefNode(node, rv)
	}

end:
	return node, found
}

func (cb *CodeBuilder) register(rv refVal, n *Node) (ref reflect.Value) {
	_, found := cb.isRegistered(rv)
	if found {
		ref = n.Ref
		goto end
	}
	cb.nodeMap[rv] = n
	n.Index = len(cb.nodeMap)
	ref = reflect.ValueOf(n.Index)
	if rv.Kind() == reflect.Pointer {
		cb.ptrMap[rv.Pointer()] = n
	}
	cb.nodes = append(cb.nodes, n)
	n.Ref = ref
end:
	return ref
}

func (cb *CodeBuilder) sortedKeys(rv refVal) (keys []reflect.Value) {
	keyValues := rv.MapKeys()
	keys = make([]reflect.Value, len(keyValues))
	for i, k := range keyValues {
		keys[i] = k
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	return keys
}

func (cb *CodeBuilder) newRefNode(node *Node, ref reflect.Value) (n *Node) {
	return NewNode(&NodeArgs{
		Name:        fmt.Sprintf("ref%d", node.Index),
		NodeRef:     node,
		Type:        RefNode,
		CodeBuilder: cb,
		Value:       ref,
		Index:       node.Index,
	})
}

// maybeReuniteNodes looks for any pointer parents with children, and/or children with
// no parents that match and connect them.
func (cb *CodeBuilder) maybeReuniteNodes() {

	pointers := filterMapFunc(cb.nodeMap, func(value reflect.Value, _ *Node) bool {
		return value.Kind() == reflect.Pointer
	})
	nonPointers := filterMapFunc(cb.nodeMap, func(value reflect.Value, _ *Node) bool {
		return value.Kind() != reflect.Pointer
	})

	for pi, p := range pointers {
		for npi, np := range nonPointers {
			if np.parent != nil {
				continue
			}
			el := pi.Elem()
			if el.Kind() != npi.Kind() {
				continue
			}
			if el.Comparable() && !el.Equal(npi) {
				continue
			}
			if !reflect.DeepEqual(el, npi) {
				continue
			}
			np.parent = p
		}
	}
}

func isSame(v1, v2 reflect.Value) (same bool) {
	if v1.Kind() == reflect.Pointer {
		v1 = v1.Elem()
	}
	if v1.Kind() != v2.Kind() {
		goto end
	}
	if v1.Comparable() && !v1.Equal(v2) {
		goto end
	}
	if !reflect.DeepEqual(v1, v2) {
		goto end
	}
	same = true
end:
	return same
}
