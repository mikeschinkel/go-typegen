package typegen_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mikeschinkel/go-typegen"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Int    int
	String string
}

type recurStruct struct {
	name  string
	recur *recurStruct
	extra string
}
type recur2Struct struct {
	recur []*recur2Struct
}
type iFaceStruct struct {
	iFace1 interface{}
	iFace2 any
}

type nM = typegen.NodeMarshaler
type nodesFunc func(m *nM) typegen.Nodes
type Node = typegen.Node
type Nodes = typegen.Nodes

var AddNode = typegen.TestAddNode
var GetNode = typegen.TestGetNode
var InitNode = typegen.TestInitNode
var FixupNodes = typegen.TestFixupNodes

type testData struct {
	name      string
	value     any
	nodes     nodesFunc
	skipNodes bool
	want      string
}

func TestNodeBuilder_Marshal(t *testing.T) {
	recur := recurStruct{name: "root", extra: "whatever"}
	recur.recur = &recur

	recur2 := recur2Struct{}
	recur2.recur = make([]*recur2Struct, 1)
	recur2.recur[0] = &recur2

	iFace := iFaceStruct{}
	iFace.iFace1 = interface{}("Hello")
	iFace.iFace2 = any(10)

	tests := []testData{
		{
			name:  "Simple string/int map",
			value: map[string]int{"Foo": 1, "Bar": 2, "Baz": 3},
			// Keys will be sorted alphabetically on output
			want:      wantValue("map[string]int", `map[string]int{"Bar":2,"Baz":3,"Foo":1,}`),
			skipNodes: true,
		},
		{
			name:      "Empty string/int map",
			value:     map[string]int{},
			want:      wantValue("map[string]int", "map[string]int{}"),
			skipNodes: true,
		},
		{
			name:      "Simple int slice",
			value:     []int{1, 2, 3},
			want:      wantValue(`[]int`, `[]int{1,2,3,}`),
			skipNodes: true,
		},
		{
			name:      "Pointer to struct with indirect property pointing to itself",
			value:     &recur2,
			want:      wantPtrValue(`recur2Struct`, `recur2Struct{recur:nil,}%s  var2 := []*recur2Struct{nil,}%s  var1.recur = var2%s  var2[0] = &var1`, "\n", "\n", "\n"),
			skipNodes: true,
		},
		{
			name:      "Pointer to struct with property pointing to itself",
			value:     &recur,
			want:      wantPtrValue(`recurStruct`, `recurStruct{name:"root",recur:nil,extra:"whatever",}%s  var1.recur = &var1`, "\n"),
			skipNodes: true,
		},
		{
			name:      "Empty array",
			value:     [0]int{},
			want:      wantValue(`[0]int`, `[0]int{}`),
			skipNodes: true,
		},
		{
			name:      "Simple int array",
			value:     [3]int{1, 2, 3},
			want:      wantValue(`[3]int`, `[3]int{1,2,3,}`),
			skipNodes: true,
		},
		{
			name:      "Simple interface containing int",
			value:     interface{}(10),
			want:      wantValue(`int`, `10`),
			skipNodes: true,
		},
		{
			name:      "Slice of `any` containing 1,2,3",
			value:     []any{1, 2, 3},
			want:      wantValue(`[]any`, `[]any{1,2,3,}`),
			skipNodes: true,
		},
		{
			name:      "Simple any slice, all same numbers",
			value:     []any{1, 1, 1},
			want:      wantValue(`[]any`, `[]any{1,1,1,}`),
			skipNodes: true,
		},
		{
			name:      "Slice of any containing \"Hello\", \"GoodBy\"",
			value:     []any{"Hello", "Goodbye"},
			want:      wantValue(`[]any`, `[]any{"Hello","Goodbye",}`),
			skipNodes: true,
		},
		{
			name:      "[]any{reflect.ValueOf(10)}",
			value:     []any{reflect.ValueOf(10)},
			want:      wantValue(`[]any`, `[]any{reflect.ValueOf(10),}`),
			skipNodes: true,
		},
		{
			name:      "Pointer to interface struct containing interface{}(string) and any(int)",
			value:     &iFace,
			want:      wantPtrValue(`iFaceStruct`, `iFaceStruct{iFace1:"Hello",iFace2:10,}`),
			skipNodes: true,
		},
		{
			name:      "Pointer to simple struct",
			value:     &testStruct{},
			want:      wantPtrValue(`testStruct`, `testStruct{Int:0,String:"",}`),
			skipNodes: true,
		},
		{
			name:      "Pointer to interface struct containing interface{}(string) and any(int)",
			value:     &iFace,
			want:      wantPtrValue(`iFaceStruct`, `iFaceStruct{iFace1:"Hello",iFace2:10,}`),
			skipNodes: true,
		},
		intNode(),
		int64Node(),
		boolNode(),
		stringNode(),
		float64Node(),
		pointerToSimpleStructNode(testStruct{}),
		emptyIntSliceNode(),
		nilNode(),
	}
	subs := typegen.Substitutions{
		reflect.TypeOf(reflect.Value{}): func(rv *reflect.Value) string {
			return fmt.Sprintf("reflect.ValueOf(%v)", rv.Interface())
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := typegen.NewNodeMarshaler(subs)
			nodes := m.Marshal(tt.value)
			//if !tt.skipNodes {
			//	assert.Equal(t, nodes, tt.nodes(m))
			//}
			b := typegen.NewCodeBuilder("getData", "typegen_test", nodes)
			got := b.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func wantValue(typ, want string, args ...any) string {
	return wantValueWithReturn(typ, want, "var1", args...)
}
func wantPtrValue(typ, want string, args ...any) string {
	if typ[0] != '*' {
		typ = "*" + typ
	}
	return wantValueWithReturn(typ, want, "&var1", args...)
}

func wantValueWithReturn(typ, want, ret string, args ...any) string {
	want = fmt.Sprintf("%s\n  return %s", want, ret)
	if len(args) > 0 {
		want = fmt.Sprintf(want, args...)
	}
	return fmt.Sprintf(`func getData() %s {%s  var1 := %s%s}`, typ, "\n", want, "\n")
}

func emptyIntSliceNode() testData {
	intSlice := make([]int, 0)
	return testData{
		name:  "Empty int slice",
		value: intSlice,
		want:  wantValue(`[]int`, `[]int{}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Value:     intSlice,
					Type:      typegen.SliceNode,
					Name:      `[]int`,
					Marshaler: m,
				},
			}, nil)
		},
	}
}
func intNode() testData {
	return testData{
		name:  "int(100)",
		value: 100,
		want:  wantValue("int", `100`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Name:      "int(100)",
					Value:     reflect.ValueOf(100),
					Type:      typegen.IntNode,
				},
			}, nil)
		},
	}
}
func boolNode() testData {
	return testData{
		name:  "Boolean true",
		value: true,
		want:  wantValue("bool", `true`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Name:      "bool(true)",
					Value:     reflect.ValueOf(true),
					Type:      typegen.BoolNode,
				},
			}, nil)
		},
	}
}
func int64Node() testData {
	return testData{
		name:  "64-bit integer",
		value: int64(100),
		want:  wantValue("int64", `int64(100)`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Name:      "int64(100)",
					Value:     reflect.ValueOf(int64(100)),
					Type:      typegen.Int64Node,
				},
			}, nil)
		},
	}
}
func float64Node() testData {
	return testData{
		name:  "Float",
		value: 1.23,
		want:  wantValue("float64", `float64(1.230000)`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Name:      "float64(1.23)",
					Value:     reflect.ValueOf(1.23),
					Type:      typegen.Float64Node,
				},
			}, nil)
		},
	}
}
func stringNode() testData {
	return testData{
		name:  "Simple String",
		value: "Hello World",
		want:  wantValue("string", `"Hello World"`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Name:      `string("Hello World")`,
					Value:     reflect.ValueOf("Hello World"),
					Type:      typegen.StringNode,
				},
			}, nil)
		},
	}
}
func pointerToSimpleStructNode(myStruct testStruct) testData {
	return testData{
		name:  "Pointer to simple struct",
		value: &myStruct,
		want:  wantPtrValue(`testStruct`, `testStruct{Int:0,String:"",}`),
		nodes: func(m *nM) Nodes {
			return FixupNodes(Nodes{
				nil,
				{
					Marshaler: m,
					Name:      "*typegen_test.testStruct",
					Value:     reflect.ValueOf(&myStruct),
					Type:      typegen.PointerNode,
				},
				{
					Marshaler: m,
					Name:      "typegen_test.testStruct",
					Value:     reflect.ValueOf(&myStruct).Elem(), // Note the .Elem()
					Type:      typegen.StructNode,
				},
			}, func(nodes typegen.Nodes) {
				for _, n := range nodes {
					InitNode(n)
				}

				nodes[2].Parent = nodes[1]

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Type:      typegen.FieldNode,
					Name:      "Int",
					Index:     0,
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 0), &Node{
					Marshaler: m,
					Name:      "int(0)",
					Type:      typegen.IntNode,
					Value:     GetNode(nodes[1], 0).Value,
					Parent:    GetNode(nodes[2], 0),
				})

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Type:      typegen.FieldNode,
					Name:      "String",
					Index:     1,
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 1), &Node{
					Marshaler: m,
					Name:      `string("")`,
					Type:      typegen.StringNode,
					Value:     GetNode(nodes[1], 1).Value,
					Parent:    GetNode(nodes[2], 1),
				})

			})

		},
	}
}
func nilNode() testData {
	return testData{
		name:  "nil",
		value: nil,
		want:  wantValue(`error`, `nil`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Value:     nil,
					Type:      typegen.InvalidNode,
					Name:      `nil`,
					Marshaler: m,
				},
			}, nil)
		},
	}
}
