package typegen

import (
	"fmt"
	"reflect"
	"sort"
)

type NodeBuilder struct {
	funcName string
	value    reflect.Value
	original any
	nodeMap  NodeMap
	nodes    Nodes
	ptrMap   PointerMap
	root     *Node
	omitPkg  string
}

func NewNodeBuilder(value any, funcName string, omitPkg string) *NodeBuilder {
	cb := &NodeBuilder{
		original: value,
		value:    reflect.ValueOf(value),
		funcName: funcName,
		omitPkg:  omitPkg,
		nodeMap:  make(NodeMap),
		ptrMap:   make(PointerMap),
		nodes:    make(Nodes, 1), // Zero element is unused so node.index==0 can represent invalid
	}

	if cb.value.Kind() == reflect.Struct {
		panic("NodeBuilder currently does not support generating code for non-pointer structs (and probably never will.) Pass a pointer to the struct instead.")
	}

	return cb
}

func (nb *NodeBuilder) Nodes() Nodes {
	return nb.nodes
}

func (nb *NodeBuilder) Build() {
	nb.root = nb.marshalValue(nb.value)

	if nb.NodeCount() == 0 {
		// If the root value is not a container, and thus not yet registered, register
		// the one value, so it can be converted in .String().
		nb.register(nb.value, nb.root)
	}

	// Ensure the root node is not duplicated if referenced elsewhere by making sure
	// all nodes are connected.
	nb.maybeReuniteNodes()
}

func (nb *NodeBuilder) NodeCount() int {
	return len(nb.nodeMap)
}

func (nb *NodeBuilder) String() string {
	return nb.Generate()
}

// selectNode selects a node for use in Generate(). If it is a pointer it
// dereferences it by selecting the next node in the list of .nodes and
// increments and returns the index. It also returns if it was a pointer
// so that Generate() can generate a pointer return value, if so.
func (nb *NodeBuilder) selectNode(index int) (n *Node, _ int, wasPtr bool) {
	n = nb.nodes[index]
	if n.Type != PointerNode {
		goto end
	}
	wasPtr = true
	if index >= nb.NodeCount() {
		goto end
	}
	index++
	n = nb.nodes[index]
end:
	return n, index, wasPtr
}

func (nb *NodeBuilder) Generate() string {
	var returnVar, returnType string
	var n *Node
	var wasPtr bool

	g := NewGenerator(nb.omitPkg)
	nodeCnt := nb.NodeCount()
	for i := 1; i <= nodeCnt; i++ {
		n, i, wasPtr = nb.selectNode(i)
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
	for _, a := range g.assignments {
		g.writeAssignment(a)
	}
	g.WriteString(fmt.Sprintf("%sreturn %s\n", g.Indent, returnVar))
	g.WriteByte('}')
	return fmt.Sprintf("func %s() %s {\n%s", nb.funcName, returnType, g.String())
}

func (nb *NodeBuilder) marshalValue(rv reflect.Value) (node *Node) {
	node = nb.marshalContainers(rv)
	if node != nil {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        "scalar",
		NodeBuilder: nb,
		Value:       rv,
	})
end:
	return node
}

func (nb *NodeBuilder) marshalContainers(rv reflect.Value) (node *Node) {

	switch rv.Kind() {
	case reflect.Ptr:
		node = nb.marshalPtr(rv)
	case reflect.Struct:
		node = nb.marshalStruct(rv)
	case reflect.Slice:
		node = nb.marshalSlice(rv)
	case reflect.Map:
		node = nb.marshalMap(rv)
	case reflect.Interface:
		node = nb.marshalInterface(rv)
	case reflect.Array:
		node = nb.marshalArray(rv)
	default:
		goto end
	}
end:
	return node
}

// marshalArray marshals an array value to create a Node
func (nb *NodeBuilder) marshalArray(rv reflect.Value) (node *Node) {
	return nb.marshalElements(rv, func() string {
		return fmt.Sprintf("[%d]%s", rv.Len(), rv.Type().Elem())
	})
}

// marshalSlice marshals a slice value to create a Node
func (nb *NodeBuilder) marshalSlice(rv reflect.Value) (node *Node) {
	return nb.marshalElements(rv, func() string {
		return fmt.Sprintf("[]%s", rv.Type().Elem())
	})
}

// marshalElements marshals both array and slice values to create Nodes
func (nb *NodeBuilder) marshalElements(rv reflect.Value, nameFunc func() string) (node *Node) {
	var ref reflect.Value

	node, found := nb.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        nameFunc(),
		NodeBuilder: nb,
		Value:       rv,
	})
	ref = nb.register(rv, node)

	node.SetNodeCount(rv.Len())
	for i := 0; i < rv.Len(); i++ {
		child := NewNode(&NodeArgs{
			Name:        fmt.Sprintf("Index %d", i),
			Type:        ElementNode,
			NodeBuilder: nb,
			Value:       reflect.ValueOf(i),
			Index:       i,
		})
		node.AddNode(child)
		childValue := nb.marshalValue(rv.Index(i))
		childValue.Index = i
		childValue.Name = fmt.Sprintf("Value %d", i)
		childValue.resetDebugString()
		child.AddNode(childValue)
	}
end:
	return nb.newRefNode(node, ref)
}

func (nb *NodeBuilder) marshalMap(rv reflect.Value) (node *Node) {
	var name string
	var ref reflect.Value

	var keys []reflect.Value

	node, found := nb.isRegistered(rv)
	if found {
		goto end
	}
	name = fmt.Sprintf("map[%s]%s", rv.Type().Key(), rv.Type().Elem())
	node = NewNode(&NodeArgs{
		Name:        name,
		NodeBuilder: nb,
		Value:       rv,
	})
	ref = nb.register(rv, node)
	keys = nb.sortedKeys(rv)
	node.SetNodeCount(len(keys))
	for _, key := range keys {
		child := nb.marshalValue(key)
		node.AddNode(child)
		child.AddNode(nb.marshalValue(rv.MapIndex(key)))
	}
end:
	return nb.newRefNode(node, ref)
}

func (nb *NodeBuilder) marshalPtr(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := nb.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:        "nil",
			NodeBuilder: nb,
			Value:       rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        "&",
		NodeBuilder: nb,
		Value:       rv,
	})
	ref = nb.register(rv, node)
	node.AddNode(nb.marshalValue(rv.Elem()))
end:
	return nb.newRefNode(node, ref)
}

func (nb *NodeBuilder) marshalInterface(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := nb.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:        "nil",
			NodeBuilder: nb,
			Value:       rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        rv.Type().String(),
		NodeBuilder: nb,
		Value:       rv,
	})
	ref = nb.register(rv, node)
	node.AddNode(nb.marshalValue(rv.Elem()))
end:
	return nb.newRefNode(node, ref)
}

func (nb *NodeBuilder) marshalStruct(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := nb.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:        rv.Type().String(),
		NodeBuilder: nb,
		Value:       rv,
	})
	ref = nb.register(rv, node)
	for i := 0; i < rv.NumField(); i++ {
		name := rv.Type().Field(i).Name
		child := NewNode(&NodeArgs{
			Name:        name,
			Type:        FieldNode,
			NodeBuilder: nb,
			Value:       reflect.ValueOf(name),
			Index:       i,
		})
		node.AddNode(child)
		child.AddNode(nb.marshalValue(rv.Field(i)))
	}
end:
	return nb.newRefNode(node, ref)
}

// register adds a Node to both .nodeMap and .nodes, and for pointers to .ptrMap.
// Used by isRegistered() to determine if a node exists or needs to be added.
// Called when marshalling collection types; array, slice, map, pointer,
// interface, and struct.
func (nb *NodeBuilder) register(rv reflect.Value, n *Node) (ref reflect.Value) {
	_, found := nb.isRegistered(rv)
	if found {
		ref = n.Ref
		goto end
	}
	nb.nodeMap[rv] = n
	n.Index = len(nb.nodeMap)
	ref = reflect.ValueOf(n.Index)
	if rv.Kind() == reflect.Pointer {
		nb.ptrMap[rv.Pointer()] = n
	}
	nb.nodes = append(nb.nodes, n)
	n.Ref = ref
end:
	return ref
}

// isRegistered returns a Node if found to be registered, and a bool true if found.
func (nb *NodeBuilder) isRegistered(rv reflect.Value) (node *Node, found bool) {

	if rv.Kind() != reflect.Pointer {
		node, found = nb.findNodeMapKey(rv)
		goto end
	}

	node, found = nb.ptrMap[rv.Pointer()]
	if found {
		// If the value of `rv` is a pointer, and we previously recorded it, then skip
		// registration.
		goto end
	}

	// Look for the value pointed to having already been registered
	node, found = nb.findNodeMapKey(rv)
	if found {
		// If yes, create a RefNode for it
		node = nb.newRefNode(node, rv)
	}

end:
	return node, found
}

// findNodeMapKey loops through NodeBuilder.nodeMap[reflect.Value]*Node and to
// find the value that matches. For pointer values it dereferences to do the
// match.
func (nb *NodeBuilder) findNodeMapKey(rv reflect.Value) (node *Node, found bool) {
	var n *Node
	var k reflect.Value

	// First look for direct reflect.Value matches
	for k, n = range nb.nodeMap {
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
	for k, n = range nb.nodeMap {
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

func (nb *NodeBuilder) sortedKeys(rv reflect.Value) (keys []reflect.Value) {
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

func (nb *NodeBuilder) newRefNode(node *Node, ref reflect.Value) (n *Node) {
	return NewNode(&NodeArgs{
		Name:        fmt.Sprintf("ref%d", node.Index),
		NodeRef:     node,
		Type:        RefNode,
		NodeBuilder: nb,
		Value:       ref,
		Index:       node.Index,
	})
}

// maybeReuniteNodes looks for any pointer parents with children, and/or children with
// no parents that match and connect them.
func (nb *NodeBuilder) maybeReuniteNodes() {

	pointers := filterMapFunc(nb.nodeMap, func(value reflect.Value, _ *Node) bool {
		return value.Kind() == reflect.Pointer
	})
	nonPointers := filterMapFunc(nb.nodeMap, func(value reflect.Value, _ *Node) bool {
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
