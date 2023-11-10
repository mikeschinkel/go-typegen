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
	var returnVar, returnType string
	g := NewGenerator(cb.omitPkg)
	nodeCnt := cb.nodes.Len()
	for i := 1; i < nodeCnt; i++ {
		n := cb.nodes[i]
		if n.Value.IsZero() {
			continue
		}
		if n.Type == PointerNode {
			if i < nodeCnt-1 {
				i++
				n = cb.nodes[i]
			}
			if returnVar == "" {
				returnVar += "&" + n.Varname()
				returnType = "*" + g.MaybeStripPackage(n.Value.Type().String())
			}
		}
		if returnVar == "" {
			returnVar = n.Varname()
			returnType = g.MaybeStripPackage(n.Value.Type().String())
		}
		g.WriteString(fmt.Sprintf("%s%s = ", g.Indent, n.Varname()))
		g.WriteCode(n)
		g.WriteByte('\n')

		// Record that this var has been generated
		g.varMap[i] = struct{}{}

		for _, a := range g.Assignments {
			g.WriteAssignment(a)
		}
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
	node = NewNode(&NodeArgs{
		Name:        name,
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
	for k, n := range cb.nodeMap {
		if !k.Equal(rv) {
			continue
		}
		node = n
		found = true
		goto end
	}
end:
	return node, found
}

func (cb *CodeBuilder) isRegistered(rv refVal) (node *Node, found bool) {
	//var index int

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
	node, found = cb.findNodeMapKey(rv.Elem())
	if found {
		// If yes, create a RefNode for it
		node = cb.newRefNode(node, rv.Elem())
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
	n.Index = cb.setVarname(n)
	n.Ref = reflect.ValueOf(n.Index)
	if rv.Kind() == reflect.Pointer {
		cb.ptrMap[rv.Pointer()] = n
	}
	cb.nodes = append(cb.nodes, n)
	ref = n.Ref
end:
	return ref
}

func (cb *CodeBuilder) setVarname(n *Node) int {
	n.Index = len(cb.nodeMap)
	return len(cb.nodeMap)
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
	n = NewNode(&NodeArgs{
		Name:        node.Varname(),
		Type:        RefNode,
		CodeBuilder: cb,
		Value:       ref,
		Index:       node.Index,
	})
	return n
}
