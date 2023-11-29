package typegen_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mikeschinkel/go-diffator"
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
var InitNodes = typegen.TestInitNodes
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
		intNode(),
		int64Node(),
		boolNode(),
		stringNode(),
		float64Node(),
		pointerToSimpleStructNode(testStruct{}),
		emptyIntSliceNode(),
		nilNode(),
		pointerToInterfaceStructContainingInterfacesNode(&iFace),
		simpleStringIntMapNode(),
		pointerToSimpleStruct(),
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
			if !tt.skipNodes {
				want := tt.nodes(m)
				got := nodes
				diff := diffator.Diff(want, got)
				if diff != "" {
					t.Errorf(diff)
				}
				//assert.Equal(t, want, got)
			}
			b := typegen.NewCodeBuilder("getData", "typegen_test", nodes)
			got := b.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func pointerToSimpleStruct() testData {
	value := &testStruct{}
	return testData{
		name:  "Pointer to simple struct",
		value: value,
		want:  wantPtrValue(`testStruct`, `testStruct{Int:0,String:"",}`),
		nodes: func(m *nM) Nodes {
			return FixupNodes(Nodes{
				nil,
				{
					Id:        1,
					Typename:  "*typegen_test.testStruct",
					Value:     value,
					Type:      typegen.PointerNode,
					Name:      `*typegen_test.testStruct`,
					Marshaler: m,
				},
				{
					Id:        2,
					Typename:  "typegen_test.testStruct",
					Value:     value,
					Type:      typegen.StructNode,
					Name:      `typegen_test.testStruct`,
					Marshaler: m,
				},
			}, func(nodes typegen.Nodes) {

				nodes = InitNodes(nodes)

				AddNode(nodes[1], nodes[2])
				nodes[2].Parent = nodes[1]
				AddNode(nodes[2], &Node{
					Marshaler: m,
					Index:     0,
					Id:        3,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "Int",
					Parent:    nodes[2],
				})
				AddNode(nodes[2], &Node{
					Marshaler: m,
					Index:     1,
					Id:        5,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "String",
					Parent:    nodes[2],
				})

				AddNode(GetNode(nodes[2], 0), &Node{
					Marshaler: m,
					Id:        4,
					Typename:  "int",
					Name:      `int(0)`,
					Type:      typegen.IntNode,
					Value:     0,
					Parent:    GetNode(nodes[2], 0),
				})

				AddNode(GetNode(nodes[2], 1), &Node{
					Marshaler: m,
					Id:        6,
					Index:     0,
					Typename:  "string",
					Name:      `string("")`,
					Type:      typegen.StringNode,
					Value:     "",
					Parent:    GetNode(nodes[2], 1),
				})

			})
		},
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
					Id:        1,
					Typename:  "[]int",
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
					Id:        1,
					Typename:  "int",
					Name:      "int(100)",
					Value:     100,
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
					Id:        1,
					Typename:  "bool",
					Name:      "bool(true)",
					Value:     true,
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
					Id:        1,
					Typename:  "int64",
					Name:      "int64(100)",
					Value:     int64(100),
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
					Id:        1,
					Typename:  "float64",
					Name:      "float64(1.23)",
					Value:     1.23,
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
					Id:        1,
					Typename:  "string",
					Name:      `string("Hello World")`,
					Value:     "Hello World",
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
					Id:        1,
					Typename:  "*typegen_test.testStruct",
					Name:      "*typegen_test.testStruct",
					Value:     &myStruct,
					Type:      typegen.PointerNode,
				},
				{
					Marshaler: m,
					Id:        2,
					Typename:  "typegen_test.testStruct",
					Name:      "typegen_test.testStruct",
					Value:     myStruct,
					Type:      typegen.StructNode,
				},
			}, func(nodes typegen.Nodes) {

				nodes = InitNodes(nodes)
				AddNode(nodes[1], nodes[2])
				nodes[2].Parent = nodes[1]

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Id:        3,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "Int",
					Index:     0,
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 0), &Node{
					Marshaler: m,
					Id:        4,
					Typename:  "int",
					Name:      "int(0)",
					Type:      typegen.IntNode,
					Value:     0,
					Parent:    GetNode(nodes[2], 0),
				})

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Id:        5,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "String",
					Index:     1,
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 1), &Node{
					Marshaler: m,
					Id:        6,
					Typename:  "string",
					Name:      `string("")`,
					Type:      typegen.StringNode,
					Value:     "",
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
					Id:        1,
					Typename:  "nil",
					Value:     nil,
					Type:      typegen.InvalidNode,
					Name:      `nil`,
					Marshaler: m,
				},
			}, nil)
		},
	}
}
func pointerToInterfaceStructContainingInterfacesNode(iFace *iFaceStruct) testData {
	return testData{
		name:  "Pointer to interface struct containing interface{}(string) and any(int)",
		value: iFace,
		want:  wantPtrValue(`iFaceStruct`, `iFaceStruct{iFace1:"Hello",iFace2:10,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Name:      `*typegen_test.iFaceStruct`,
					Type:      typegen.PointerNode,
					Typename:  "*typegen_test.iFaceStruct",
					Value:     iFace,
				},
				{
					Marshaler: m,
					Id:        2,
					Name:      `typegen_test.iFaceStruct`,
					Type:      typegen.StructNode,
					Typename:  "typegen_test.iFaceStruct",
					Value:     *iFace,
				},
				{
					Marshaler: m,
					Id:        4,
					Index:     0,
					Name:      `any("Hello")`,
					Type:      typegen.InterfaceNode,
					Typename:  "any(string)",
					Value:     "Hello",
				},
				{
					Marshaler: m,
					Id:        7,
					Index:     1,
					Name:      `any(10)`,
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     10,
				},
			}, func(nodes typegen.Nodes) {

				nodes = InitNodes(nodes)

				AddNode(nodes[1], nodes[2])
				nodes[2].Parent = nodes[1]
				AddNode(nodes[2], &Node{
					Marshaler: m,
					Index:     0,
					Id:        3,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "iFace1",
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 0), nodes[3])
				AddNode(nodes[2], &Node{
					Marshaler: m,
					Index:     1,
					Id:        6,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "iFace2",
					Parent:    nodes[2],
				})
				AddNode(GetNode(nodes[2], 1), nodes[4])

				AddNode(nodes[3], &Node{
					Marshaler: m,
					Id:        5,
					Index:     0,
					Typename:  "string",
					Name:      `string("Hello")`,
					Type:      typegen.StringNode,
					Value:     "Hello",
					Parent:    nodes[3],
				})

				AddNode(nodes[4], &Node{
					Marshaler: m,
					Id:        8,
					Typename:  "int",
					Name:      `int(10)`,
					Type:      typegen.IntNode,
					Value:     10,
					Parent:    nodes[4],
				})

			})
		},
	}

}
func simpleStringIntMapNode() testData {
	intMap := map[string]int{"Foo": 1, "Bar": 2, "Baz": 3}
	return testData{
		name:  "Simple string/int map",
		value: intMap,
		// Keys will be sorted alphabetically on output
		want:      wantValue("map[string]int", `map[string]int{"Bar":2,"Baz":3,"Foo":1,}`),
		skipNodes: false,
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Id:        1,
					Typename:  "map[string]int",
					Value:     intMap,
					Type:      typegen.MapNode,
					Name:      `map[string]int`,
					Marshaler: m,
				},
			}, func(nodes typegen.Nodes) {

				nodes = InitNodes(nodes)

				AddNode(nodes[1], &Node{
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.StringNode,
					Name:      `string("Bar")`,
					Parent:    nodes[1],
					Typename:  "string",
					Value:     "Bar",
				})
				AddNode(GetNode(nodes[1], 0), &Node{
					Marshaler: m,
					Id:        3,
					Type:      typegen.IntNode,
					Name:      "int(2)",
					Parent:    GetNode(nodes[1], 0),
					Typename:  "int",
					Value:     2,
				})
				AddNode(nodes[1], &Node{
					Marshaler: m,
					Id:        4,
					Index:     1,
					Type:      typegen.StringNode,
					Name:      `string("Baz")`,
					Parent:    nodes[1],
					Typename:  "string",
					Value:     "Baz",
				})
				AddNode(GetNode(nodes[1], 1), &Node{
					Marshaler: m,
					Id:        5,
					Type:      typegen.IntNode,
					Name:      "int(3)",
					Parent:    GetNode(nodes[1], 1),
					Typename:  "int",
					Value:     3,
				})
				AddNode(nodes[1], &Node{
					Marshaler: m,
					Id:        6,
					Index:     2,
					Type:      typegen.StringNode,
					Name:      `string("Foo")`,
					Parent:    nodes[1],
					Typename:  "string",
					Value:     "Foo",
				})
				AddNode(GetNode(nodes[1], 2), &Node{
					Marshaler: m,
					Id:        7,
					Type:      typegen.IntNode,
					Name:      "int(1)",
					Parent:    GetNode(nodes[1], 2),
					Typename:  "int",
					Value:     1,
				})
			})
		},
	}
}
