package typegen_test

import (
	"github.com/mikeschinkel/go-typegen"
)

func emptyIntSliceNode() testData {
	intSlice := make([]int, 0)
	return testData{
		name:  "Empty int slice",
		value: intSlice,
		want:  wantValue(`[]int`, `[]int{}`),
		nodes: func(m *nM) Nodes {
			return FixupNodes(Nodes{
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
