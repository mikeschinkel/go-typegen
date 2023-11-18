package typegen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type NodeMarshaler struct {
	value       reflect.Value
	original    any
	nodeMap     NodeMap
	nodes       Nodes
	ptrMap      PointerMap
	root        *Node
	debugString string
}

func NewNodeMarshaler(value any) *NodeMarshaler {
	m := &NodeMarshaler{
		original: value,
		value:    reflect.ValueOf(value),
		nodeMap:  make(NodeMap),
		ptrMap:   make(PointerMap),
		nodes:    make(Nodes, 1), // Zero element is unused so node.index==0 can represent invalid
	}

	if m.value.Kind() == reflect.Struct {
		panic("NodeMarshaler currently does not support generating code for non-pointer structs (and probably never will.) Pass a pointer to the struct instead.")
	}

	m.resetDebugString()

	return m
}

func (m *NodeMarshaler) Nodes() Nodes {
	return m.nodes
}

func (m *NodeMarshaler) Marshal() Nodes {
	m.root = m.marshalValue(m.value)

	if m.NodeCount() == 0 {
		// If the root value is not a container, and thus not yet registered, registerNode
		// the one value, so it can be converted in .String().
		m.registerNode(m.value, m.root)
	}

	// Ensure the root node is not duplicated if referenced elsewhere by making sure
	// all nodes are connected.
	m.maybeReuniteNodes()

	return m.nodes
}

func (m *NodeMarshaler) NodeCount() int {
	return len(m.nodeMap)
}

func (m *NodeMarshaler) marshalValue(rv reflect.Value) (node *Node) {
	node = m.marshalContainers(rv)
	if node != nil {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:      "scalar",
		marshaler: m,
		Value:     rv,
	})
end:
	return node
}

func (m *NodeMarshaler) marshalContainers(rv reflect.Value) (node *Node) {

	switch rv.Kind() {
	case reflect.Ptr:
		node = m.marshalPtr(rv)
	case reflect.Struct:
		node = m.marshalStruct(rv)
	case reflect.Slice:
		node = m.marshalSlice(rv)
	case reflect.Map:
		node = m.marshalMap(rv)
	case reflect.Interface:
		node = m.marshalInterface(rv)
	case reflect.Array:
		node = m.marshalArray(rv)
	default:
		goto end
	}
end:
	return node
}

// marshalArray marshals an array value to create a Node
func (m *NodeMarshaler) marshalArray(rv reflect.Value) (node *Node) {
	return m.marshalElements(rv, func() string {
		return fmt.Sprintf("[%d]%s", rv.Len(), rv.Type().Elem())
	})
}

// marshalSlice marshals a slice value to create a Node
func (m *NodeMarshaler) marshalSlice(rv reflect.Value) (node *Node) {
	return m.marshalElements(rv, func() string {
		return fmt.Sprintf("[]%s", rv.Type().Elem())
	})
}

// marshalElements marshals both array and slice values to create Nodes
func (m *NodeMarshaler) marshalElements(rv reflect.Value, nameFunc func() string) (node *Node) {
	var ref reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:      nameFunc(),
		marshaler: m,
		Value:     rv,
	})
	ref = m.registerNode(rv, node)

	node.SetNodeCount(rv.Len())
	for i := 0; i < rv.Len(); i++ {
		child := NewNode(&NodeArgs{
			Name:      fmt.Sprintf("Index %d", i),
			Type:      ElementNode,
			marshaler: m,
			Value:     reflect.ValueOf(i),
			Index:     i,
		})
		node.AddNode(child)
		childValue := m.marshalValue(rv.Index(i))
		childValue.Index = i
		childValue.Name = fmt.Sprintf("Value %d", i)
		childValue.resetDebugString()
		child.AddNode(childValue)
	}
end:
	return m.newRefNode(node, ref)
}

func (m *NodeMarshaler) marshalMap(rv reflect.Value) (node *Node) {
	var name string
	var ref reflect.Value

	var keys []reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	name = fmt.Sprintf("map[%s]%s", rv.Type().Key(), rv.Type().Elem())
	node = NewNode(&NodeArgs{
		Name:      name,
		marshaler: m,
		Value:     rv,
	})
	ref = m.registerNode(rv, node)
	keys = m.sortedKeys(rv)
	node.SetNodeCount(len(keys))
	for _, key := range keys {
		child := m.marshalValue(key)
		node.AddNode(child)
		child.AddNode(m.marshalValue(rv.MapIndex(key)))
	}
end:
	return m.newRefNode(node, ref)
}

func (m *NodeMarshaler) marshalPtr(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:      "nil",
			marshaler: m,
			Value:     rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:      "&",
		marshaler: m,
		Value:     rv,
	})
	ref = m.registerNode(rv, node)
	node.AddNode(m.marshalValue(rv.Elem()))
end:
	return m.newRefNode(node, ref)
}

func (m *NodeMarshaler) marshalInterface(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	if rv.IsNil() {
		node = NewNode(&NodeArgs{
			Name:      "nil",
			marshaler: m,
			Value:     rv,
		})
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:      rv.Type().String(),
		marshaler: m,
		Value:     rv,
	})
	ref = m.registerNode(rv, node)
	node.AddNode(m.marshalValue(rv.Elem()))
end:
	return m.newRefNode(node, ref)
}

func (m *NodeMarshaler) marshalStruct(rv reflect.Value) (node *Node) {
	var ref reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	node = NewNode(&NodeArgs{
		Name:      rv.Type().String(),
		marshaler: m,
		Value:     rv,
	})
	ref = m.registerNode(rv, node)
	for i := 0; i < rv.NumField(); i++ {
		name := rv.Type().Field(i).Name
		child := NewNode(&NodeArgs{
			Name:      name,
			Type:      FieldNode,
			marshaler: m,
			Value:     reflect.ValueOf(name),
			Index:     i,
		})
		node.AddNode(child)
		child.AddNode(m.marshalValue(rv.Field(i)))
	}
end:
	return m.newRefNode(node, ref)
}

// register adds a Node to both .nodeMap and .nodes, and for pointers to .ptrMap.
// Used by isRegistered() to determine if a node exists or needs to be added.
// Called when marshalling collection types; array, slice, map, pointer,
// interface, and struct.
func (m *NodeMarshaler) registerNode(rv reflect.Value, n *Node) (ref reflect.Value) {
	_, found := m.isRegistered(rv)
	if found {
		ref = n.Ref
		goto end
	}
	m.nodeMap[rv] = n
	n.Index = len(m.nodeMap)
	ref = reflect.ValueOf(n.Index)
	if rv.Kind() == reflect.Pointer {
		m.ptrMap[rv.Pointer()] = n
	}
	m.nodes = append(m.nodes, n)
	n.Ref = ref
	m.resetDebugString()
end:
	return ref
}

// isRegistered returns a Node if found to be registered, and a bool true if found.
func (m *NodeMarshaler) isRegistered(rv reflect.Value) (node *Node, found bool) {

	if rv.Kind() != reflect.Pointer {
		node, found = m.findNodeMapKey(rv)
		goto end
	}

	node, found = m.ptrMap[rv.Pointer()]
	if found {
		// If the value of `rv` is a pointer, and we previously recorded it, then skip
		// registration.
		goto end
	}

	// Look for the value pointed to having already been registered
	node, found = m.findNodeMapKey(rv)
	if found {
		// If yes, create a RefNode for it
		node = m.newRefNode(node, rv)
	}

end:
	return node, found
}

// findNodeMapKey loops through nodeBuilder.nodeMap[reflect.Value]*Node and to
// find the value that matches. For pointer values it dereferences to do the
// match.
func (m *NodeMarshaler) findNodeMapKey(rv reflect.Value) (node *Node, found bool) {
	var n *Node
	var k reflect.Value

	// First look for direct reflect.Value matches
	for k, n = range m.nodeMap {
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
	for k, n = range m.nodeMap {
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

func (m *NodeMarshaler) sortedKeys(rv reflect.Value) (keys []reflect.Value) {
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

func (m *NodeMarshaler) newRefNode(node *Node, ref reflect.Value) (n *Node) {
	return NewNode(&NodeArgs{
		Name:      fmt.Sprintf("ref%d", node.Index),
		NodeRef:   node,
		Type:      RefNode,
		marshaler: m,
		Value:     ref,
		Index:     node.Index,
	})
}

// maybeReuniteNodes looks for any pointer parents with children, and/or children with
// no parents that match and connect them.
func (m *NodeMarshaler) maybeReuniteNodes() {

	pointers := filterMapFunc(m.nodeMap, func(value reflect.Value, _ *Node) bool {
		return value.Kind() == reflect.Pointer
	})
	nonPointers := filterMapFunc(m.nodeMap, func(value reflect.Value, _ *Node) bool {
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

func (m *NodeMarshaler) resetDebugString() {
	sb := strings.Builder{}
	for index := len(m.nodes) - 1; index >= 1; index-- {
		sb.WriteByte(' ')
		sb.WriteString(m.nodes[index].Type.String())
	}
	m.debugString = fmt.Sprintf("[%d]%s", len(m.nodeMap), sb.String())
}
