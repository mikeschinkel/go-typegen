package typegen

import (
	"fmt"
	"reflect"
	"sort"
)

type refVal = reflect.Value

type NodeMap map[reflect.Value]*Node
type PointerMap map[uintptr]*Node

func (m NodeMap) Len() int {
	return len(m)
}

type CodeBuilder struct {
	funcName string
	value    reflect.Value
	original any
	nodeMap  NodeMap
	nodes    Nodes
	ptrMap   PointerMap
	root     *Node
}

func NewCodeBuilder(value any, funcName string) *CodeBuilder {
	cb := &CodeBuilder{
		original: value,
		value:    reflect.ValueOf(value),
		funcName: funcName,
		nodeMap:  make(NodeMap),
		ptrMap:   make(PointerMap),
		nodes:    make(Nodes, 0),
	}

	//// If the value is a pointer to an interface{}, we want to get the interface.
	//// Then we can use .Elem() to get to the concrete value that the interface holds.
	//switch rv.Kind() {
	//case reflect.Interface:
	//	rv = rv.Elem() // Just dereference the interface.
	//case reflect.Ptr:
	//	if rv.Elem().Kind() == reflect.Interface {
	//		rv = rv.Elem().Elem() // Twice .Elem(): one to dereference the pointer, one to get the interface value.
	//	}
	//}
	return cb
}

func (cb *CodeBuilder) Nodes() Nodes {
	return cb.nodes
}

func (cb *CodeBuilder) Build() {
	cb.root = cb.marshalValue(cb.value)
}

func (cb *CodeBuilder) String() string {
	g := NewGenerator()
	g.WriteString(fmt.Sprintf("func %s() string {\n", cb.funcName))
	for i := 0; i < len(cb.nodes); i++ {
		n := cb.nodes[i]
		if n.Value.IsZero() {
			continue
		}
		//index = cb.nodes.Len() - 1
		//if index >= 0 {
		//	prior := cb.nodes[index]
		//	if prior.Type == PointerNode && prior.Value.Elem() == rv {
		//		// If prior was a pointer, and it was pointing to value of `rv` then skip
		//		// registration.
		//		node= prior
		//		found = true
		//		goto end
		//	}
		//}

		g.WriteString(fmt.Sprintf("%svar%d := ", g.Indent, i))
		n.Index = i
		g.WriteCode(n)
		g.WriteByte('\n')
	}
	g.WriteByte('}')
	return g.String()
}

func (cb *CodeBuilder) marshalValue(rv refVal) (node *Node) {
	node = cb.marshalContainers(rv)
	if node != nil {
		goto end
	}
	node = NewNodeWithValue(cb, "scalar", rv)
end:
	return node
}

func (cb *CodeBuilder) marshalContainers(rv refVal) (node *Node) {

	switch rv.Kind() {
	case reflect.Struct:
		node = cb.marshalStruct(rv)
	case reflect.Slice:
		node = cb.marshalSlice(rv)
	case reflect.Map:
		node = cb.marshalMap(rv)
	case reflect.Ptr:
		node = cb.marshalPtr(rv)
	case reflect.Interface:
		node = cb.marshalInterface(rv)
	default:
		goto end
	}
end:
	return node
}

func (cb *CodeBuilder) marshalSlice(rv refVal) (node *Node) {
	var name string
	var ref refVal

	node, found := cb.isRegistered(rv)
	if found {
		goto end
	}
	name = fmt.Sprintf("[]%s", rv.Type().Elem())
	node = NewNodeWithValue(cb, name, rv)
	ref = cb.register(rv, node)

	node.SetNodeCount(rv.Len())
	for i := 0; i < rv.Len(); i++ {
		child := NewTypedNodeWithValue(cb, fmt.Sprintf("Index %d", i), ElementNode, reflect.ValueOf(i))
		node.AddNode(child)
		childValue := cb.marshalValue(rv.Index(i))
		childValue.Name = fmt.Sprintf("Value %d", i)
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
	node = NewNodeWithValue(cb, name, rv)
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
		node = NewNodeWithValue(cb, "nil", rv)
		goto end
	}
	node = NewNodeWithValue(cb, "&", rv)
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
		node = NewNodeWithValue(cb, "nil", rv)
		goto end
	}
	node = NewNodeWithValue(cb, rv.Type().String(), rv)
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
	node = NewNodeWithValue(cb, rv.Type().String(), rv)
	ref = cb.register(rv, node)
	for i := 0; i < rv.NumField(); i++ {
		name := rv.Type().Field(i).Name
		node.AddNode(NewTypedNodeWithValue(cb, name, FieldNameNode, reflect.ValueOf(name)).
			AddNode(cb.marshalValue(rv.Field(i))))
	}
end:
	return cb.newRefNode(node, ref)
}

func (cb *CodeBuilder) isRegistered(rv refVal) (node *Node, found bool) {
	//var index int

	if rv.Kind() == reflect.Pointer {
		node, found = cb.ptrMap[rv.Pointer()]
		if found {
			// If the value of `rv` is a pointer, and we previously recorded it, then skip
			// registration.
			goto end
		}
	}
	node, found = cb.nodeMap[rv]
end:
	return node, found
}

func (cb *CodeBuilder) register(rv refVal, n *Node) (ref reflect.Value) {
	_, found := cb.isRegistered(rv)
	if found {
		ref = n.Ref
		goto end
	}
	n.Index = cb.setVarname(n)
	n.Ref = reflect.ValueOf(n.Index)
	cb.nodeMap[rv] = n
	if rv.Kind() == reflect.Pointer {
		cb.ptrMap[rv.Pointer()] = n
	}
	cb.nodes = append(cb.nodes, n)
	ref = n.Ref
end:
	return ref
}

func (cb *CodeBuilder) setVarname(n *Node) int {
	n.Varname = fmt.Sprintf("var%d", len(cb.nodeMap))
	return len(cb.nodeMap)
}

func (cb *CodeBuilder) registeredNode(rv refVal) *Node {
	n, _ := cb.nodeMap[rv]
	return n
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
	if !ref.IsValid() {
		return node
	}
	return NewTypedNodeWithValue(cb,
		fmt.Sprintf("var%s", node.Varname[3:]),
		RefNode,
		ref)
}
