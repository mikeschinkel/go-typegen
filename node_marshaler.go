package typegen

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/mikeschinkel/go-typegen/ezreflect"
)

type Substitutions map[reflect.Type]func(*reflect.Value) string

type NodeMarshaler struct {
	original      any
	nodeMap       NodeMap
	nodes         Nodes
	ptrMap        PointerMap
	root          *Node
	debugString   string
	substitutions Substitutions
	nextNodeId    int
}

func NewNodeMarshaler(subs Substitutions) *NodeMarshaler {
	m := &NodeMarshaler{
		substitutions: subs,
	}
	resetDebugString(m)

	return m
}

func (m *NodeMarshaler) NewNode(args *NodeArgs) (n *Node) {
	m.nextNodeId++
	return NewNode(m.nextNodeId, args)
}

func (m *NodeMarshaler) reinitialize() {
	m.nodeMap = make(NodeMap)
	m.ptrMap = make(PointerMap)
	// Zero element is unused so node.index==0 can represent invalid
	m.nodes = make(Nodes, 1)
}

func (m *NodeMarshaler) Nodes() Nodes {
	return m.nodes
}

func (m *NodeMarshaler) Marshal(value any) Nodes {
	m.reinitialize()
	rv := reflect.ValueOf(value)

	if rv.Kind() == reflect.Struct {
		panic("NodeMarshaler.Marshal() currently does not support generating code " +
			"for non-pointer structs (and probably never will.) " +
			"Pass a pointer to the struct instead.",
		)
	}

	m.original = value
	m.root = m.marshalValue(&rv, nil)

	if m.NodeCount() == 0 {
		// If the root value is not a container, and thus not yet registered, registerNode
		// the one value, so it can be converted in .String().
		m.registerNode(&rv, m.root)
	}

	// Ensure the root node is not duplicated if referenced elsewhere by making sure
	// all nodes are connected.
	//m.maybeReuniteNodes()

	return m.nodes
}

func (m *NodeMarshaler) NodeCount() int {
	return len(m.nodeMap)
}

func (m *NodeMarshaler) marshalValue(rv *reflect.Value, parent *Node) (node *Node) {
	var name string
	if rv.IsValid() {
		subsFunc, ok := m.substitutions[rv.Type()]
		if ok {
			reflector := reflect.ValueOf(subsFunc(rv))
			node = m.NewNode(&NodeArgs{
				Name:         "substitution",
				marshaler:    m,
				Type:         SubstitutionNode,
				ReflectValue: &reflector,
				Parent:       parent,
			})
			goto end
		}
	}
	node = m.marshalContainers(rv, parent)
	if node != nil {
		goto end
	}
	name = "nil"
	if rv.IsValid() {
		name = fmt.Sprintf("%s(%s)",
			rv.Type().String(),
			ezreflect.NewReflectWrapper(*rv).String(),
		)
	}
	// Scalar
	node = m.NewNode(&NodeArgs{
		Name:         name,
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
end:
	return node
}

func (m *NodeMarshaler) marshalContainers(rv *reflect.Value, parent *Node) (node *Node) {

	switch rv.Kind() {
	case reflect.Ptr:
		node = m.marshalPointer(rv, parent)
	case reflect.Struct:
		node = m.marshalStruct(rv, parent)
	case reflect.Slice:
		node = m.marshalSlice(rv, parent)
	case reflect.Map:
		node = m.marshalMap(rv, parent)
	case reflect.Interface:
		node = m.marshalInterface(rv, parent)
	case reflect.Array:
		node = m.marshalArray(rv, parent)
	default:
		goto end
	}
end:
	return node
}

// marshalArray marshals an array value to create a Node
func (m *NodeMarshaler) marshalArray(rv *reflect.Value, parent *Node) (node *Node) {
	return m.marshalElements(rv, parent, func() string {
		return fmt.Sprintf("[%d]%s", rv.Len(), rv.Type().Elem())
	})
}

// marshalSlice marshals a slice value to create a Node
func (m *NodeMarshaler) marshalSlice(rv *reflect.Value, parent *Node) (node *Node) {
	return m.marshalElements(rv, parent, func() string {
		return fmt.Sprintf("[]%s", rv.Type().Elem())
	})
}

// marshalElements marshals both array and slice values to create Nodes
func (m *NodeMarshaler) marshalElements(rv *reflect.Value, parent *Node, nameFunc func() string) (node *Node) {
	var index reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	node = m.NewNode(&NodeArgs{
		Name:         nameFunc(),
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
	m.registerNode(rv, node)

	node.SetNodeCount(rv.Len())
	for i := 0; i < rv.Len(); i++ {
		reflector := reflect.ValueOf(i)
		child := m.NewNode(&NodeArgs{
			Name:         fmt.Sprintf("Index %d", i),
			Type:         ElementNode,
			marshaler:    m,
			ReflectValue: &reflector,
			Index:        i,
		})
		node.AddNode(child)
		index = rv.Index(i)
		childValue := m.marshalValue(&index, node)
		childValue.Name = fmt.Sprintf("Value %d", i)
		resetDebugString(childValue)
		child.AddNode(childValue)
	}
end:
	return node
}

func (m *NodeMarshaler) marshalMap(rv *reflect.Value, parent *Node) (node *Node) {
	var name string
	var index reflect.Value

	var keys []reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	name = fmt.Sprintf("map[%s]%s", rv.Type().Key(), rv.Type().Elem())
	node = m.NewNode(&NodeArgs{
		Name:         name,
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
	m.registerNode(rv, node)
	keys = m.sortedKeys(rv)
	node.SetNodeCount(len(keys))
	for _, key := range keys {
		child := m.marshalValue(&key, node)
		node.AddNode(child)
		index = rv.MapIndex(key)
		child.AddNode(m.marshalValue(&index, child))
	}
end:
	return node
}

func (m *NodeMarshaler) marshalPointer(rv *reflect.Value, parent *Node) (node *Node) {
	var name string
	var elem reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	name = rv.Type().String()
	if rv.IsNil() {
		node = m.NewNode(&NodeArgs{
			Name:         name + " (nil)",
			marshaler:    m,
			ReflectValue: rv,
			Parent:       parent,
		})
		goto end
	}
	node = m.NewNode(&NodeArgs{
		Name:         name,
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
	m.registerNode(rv, node)
	elem = rv.Elem()
	node.AddNode(m.marshalValue(&elem, node))
end:
	return node
}

func (m *NodeMarshaler) marshalInterface(rv *reflect.Value, parent *Node) (node *Node) {
	var name string
	var elem reflect.Value

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	name = m.asString(rv)
	if rv.IsNil() {
		node = m.NewNode(&NodeArgs{
			Name:         name + " (nil)",
			marshaler:    m,
			ReflectValue: rv,
			Parent:       parent,
		})
		goto end
	}
	node = m.NewNode(&NodeArgs{
		Name:         name,
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
	m.registerNode(rv, node)
	elem = rv.Elem()
	node.AddNode(m.marshalValue(&elem, node))
end:
	return node
}

func (m *NodeMarshaler) asString(rv *reflect.Value) (s string) {
	return ezreflect.AsString(*rv)
}

func (m *NodeMarshaler) marshalStruct(rv *reflect.Value, parent *Node) (node *Node) {
	var rt reflect.Type

	node, found := m.isRegistered(rv)
	if found {
		goto end
	}
	node = m.NewNode(&NodeArgs{
		Name:         rv.Type().String(),
		marshaler:    m,
		ReflectValue: rv,
		Parent:       parent,
	})
	m.registerNode(rv, node)
	rt = rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		child := m.NewNode(&NodeArgs{
			Name:      rt.Field(i).Name,
			Type:      FieldNode,
			marshaler: m,
			Index:     i,
			Typename:  "field", // TODO Decide something better, maybe?
		})
		node.AddNode(child)
		crv := rv.Field(i)
		child.AddNode(m.marshalValue(&crv, child))
	}
end:
	return node
}

// register adds a Node to both .nodeMap and .nodes, and for pointers to .ptrMap.
// Used by isRegistered() to determine if a node exists or needs to be added.
// Called when marshalling collection types; array, slice, map, pointer,
// interface, and struct.
func (m *NodeMarshaler) registerNode(rv *reflect.Value, n *Node) {
	_, found := m.isRegistered(rv)
	if found {
		goto end
	}
	m.nodeMap[*rv] = n
	resetDebugString(n)
	if rv.Kind() == reflect.Pointer {
		m.ptrMap[rv.Pointer()] = n
	}
	m.nodes = append(m.nodes, n)
	resetDebugString(m)
end:
}

// isRegistered returns a Node if found to be registered, and a bool true if found.
func (m *NodeMarshaler) isRegistered(rv *reflect.Value) (node *Node, found bool) {

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

end:
	return node, found
}

// findNodeMapKey loops through nodeBuilder.nodeMap[reflect.Value]*Node and to
// find the value that matches. For pointer values it dereferences to do the
// match.
func (m *NodeMarshaler) findNodeMapKey(rv *reflect.Value) (node *Node, found bool) {
	var n *Node
	var k reflect.Value

	// First look for direct reflect.Value matches
	for k, n = range m.nodeMap {
		if rv.Kind() == reflect.Pointer {
			continue
		}
		if !ezreflect.Equivalent(k, rv) {
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
		if !ezreflect.Equivalent(k, rv.Elem()) {
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

func (m *NodeMarshaler) sortedKeys(rv *reflect.Value) (keys []reflect.Value) {
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
