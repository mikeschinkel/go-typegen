package typegen_test

import (
	"reflect"

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
func int100Node() testData {
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
func pointerToSimpleStructNode() testData {
	type testStruct struct {
		Int    int
		String string
	}
	myStruct := testStruct{}
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

type IFaceStruct struct {
	iFace1 interface{}
	iFace2 any
}

func pointerToInterfaceStructContainingInterfacesNode() testData {

	iFace := IFaceStruct{}
	iFace.iFace1 = interface{}("Hello")
	iFace.iFace2 = any(10)

	return testData{
		name:  "Pointer to interface struct containing interface{}(string) and any(int)",
		value: &iFace,
		want:  wantPtrValue(`IFaceStruct`, `IFaceStruct{iFace1:"Hello",iFace2:10,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Name:      `*typegen_test.IFaceStruct`,
					Type:      typegen.PointerNode,
					Typename:  "*typegen_test.IFaceStruct",
					Value:     &iFace,
				},
				{
					Marshaler: m,
					Id:        2,
					Name:      `typegen_test.IFaceStruct`,
					Type:      typegen.StructNode,
					Typename:  "typegen_test.IFaceStruct",
					Value:     iFace,
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

				AddNode(nodes[1], nodes[2])
				nodes[2].Parent = nodes[1]
				AddNode(nodes[2], &Node{
					Parent:    nodes[2],
					Marshaler: m,
					Index:     0,
					Id:        3,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "iFace1",
				})
				AddNode(GetNode(nodes[2], 0), nodes[3])
				nodes[3].Parent = GetNode(nodes[2], 0)

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
				nodes[4].Parent = GetNode(nodes[2], 1)

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
	type testStruct struct {
		Int    int
		String string
	}
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
					Value:     *value,
					Type:      typegen.StructNode,
					Name:      `typegen_test.testStruct`,
					Marshaler: m,
				},
			}, func(nodes typegen.Nodes) {

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
func sliceOfAnyContainingHelloGoodbye() testData {
	value := []any{"Hello", "Goodbye"}
	return testData{
		name:  "Slice of any containing \"Hello\", \"GoodBye\"",
		value: value,
		want:  wantValue(`[]any`, `[]any{"Hello","Goodbye",}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Index:     0,
					Type:      typegen.SliceNode,
					Value:     value,
					Name:      "[]interface {}",
					Typename:  "[]interface {}",
				},
				{
					Marshaler: m,
					Index:     0,
					Id:        3,
					Type:      typegen.InterfaceNode,
					Name:      `Value 0`,
					Typename:  "any(string)",
					Value:     `Hello`,
				},
				{
					Marshaler: m,
					Index:     0,
					Id:        6,
					Type:      typegen.InterfaceNode,
					Name:      `Value 1`,
					Typename:  "any(string)",
					Value:     `Goodbye`,
				},
			}, func(nodes typegen.Nodes) {

				AddNode(nodes[1], &Node{
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.ElementNode,
					Name:      `Index 0`,
					Parent:    nodes[1],
					Typename:  "element",
					Value:     0,
				})
				AddNode(GetNode(nodes[1], 0), nodes[2])

				nodes[2].Parent = GetNode(nodes[1], 0)
				AddNode(nodes[2], &Node{
					Marshaler: m,
					Index:     0,
					Id:        4,
					Type:      typegen.StringNode,
					Name:      `string("Hello")`,
					Parent:    nodes[2],
					Typename:  "string",
					Value:     `Hello`,
				})

				AddNode(nodes[1], &Node{
					Marshaler: m,
					Index:     1,
					Id:        5,
					Type:      typegen.ElementNode,
					Name:      `Index 1`,
					Parent:    nodes[1],
					Typename:  "element",
					Value:     1,
				})
				AddNode(GetNode(nodes[1], 1), nodes[3])

				nodes[3].Parent = GetNode(nodes[1], 1)
				AddNode(nodes[3], &Node{
					Marshaler: m,
					Index:     0,
					Id:        7,
					Type:      typegen.StringNode,
					Name:      `string("Goodbye")`,
					Parent:    nodes[3],
					Typename:  "string",
					Value:     "Goodbye",
				})
			})
		},
	}
}
func simpleAnySliceAllSameNumbers() testData {
	value := []any{1, 1, 1}
	return testData{
		name:  "Simple any slice, all same numbers",
		value: value,
		want:  wantValue(`[]any`, `[]any{1,1,1,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.SliceNode,
					Name:      "[]interface {}",
					Typename:  "[]interface {}",
					Value:     value,
				},
				{
					Marshaler: m,
					Id:        3,
					Name:      "Value 0",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     1,
				},
				{
					Marshaler: m,
					Id:        6,
					Name:      "Value 1",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     1,
				},
				{
					Marshaler: m,
					Id:        9,
					Name:      "Value 2",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     1,
				},
			}, func(nodes typegen.Nodes) {

				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.ElementNode,
					Name:      `Index 0`,
					Typename:  "element",
					Value:     0,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     1,
					Id:        5,
					Type:      typegen.ElementNode,
					Name:      `Index 1`,
					Typename:  "element",
					Value:     1,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     2,
					Id:        8,
					Type:      typegen.ElementNode,
					Name:      `Index 2`,
					Typename:  "element",
					Value:     2,
				})

				nodes[2].Parent = GetNode(nodes[1], 0)
				nodes[3].Parent = GetNode(nodes[1], 1)
				nodes[4].Parent = GetNode(nodes[1], 2)

				AddNode(nodes[2], &Node{
					Parent:    nodes[2],
					Id:        4,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(1)`,
					Typename:  "int",
					Value:     1,
				})
				AddNode(nodes[3], &Node{
					Parent:    nodes[3],
					Id:        7,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(1)`,
					Typename:  "int",
					Value:     1,
				})
				AddNode(nodes[4], &Node{
					Parent:    nodes[4],
					Id:        10,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(1)`,
					Typename:  "int",
					Value:     1,
				})

				AddNode(GetNode(nodes[1], 0), nodes[2])
				AddNode(GetNode(nodes[1], 1), nodes[3])
				AddNode(GetNode(nodes[1], 2), nodes[4])

			})
		},
	}
}
func simpleAnySlice123() testData {
	value := []any{1, 2, 3}
	return testData{
		name:  "Slice of `any` containing 1,2,3",
		value: value,
		want:  wantValue(`[]any`, `[]any{1,2,3,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.SliceNode,
					Name:      "[]interface {}",
					Typename:  "[]interface {}",
					Value:     value,
				},
				{
					Marshaler: m,
					Id:        3,
					Name:      "Value 0",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     1,
				},
				{
					Marshaler: m,
					Id:        6,
					Name:      "Value 1",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     2,
				},
				{
					Marshaler: m,
					Id:        9,
					Name:      "Value 2",
					Type:      typegen.InterfaceNode,
					Typename:  "any(int)",
					Value:     3,
				},
			}, func(nodes typegen.Nodes) {

				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.ElementNode,
					Name:      `Index 0`,
					Typename:  "element",
					Value:     0,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     1,
					Id:        5,
					Type:      typegen.ElementNode,
					Name:      `Index 1`,
					Typename:  "element",
					Value:     1,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     2,
					Id:        8,
					Type:      typegen.ElementNode,
					Name:      `Index 2`,
					Typename:  "element",
					Value:     2,
				})

				nodes[2].Parent = GetNode(nodes[1], 0)
				nodes[3].Parent = GetNode(nodes[1], 1)
				nodes[4].Parent = GetNode(nodes[1], 2)

				AddNode(nodes[2], &Node{
					Parent:    nodes[2],
					Id:        4,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(1)`,
					Typename:  "int",
					Value:     1,
				})
				AddNode(nodes[3], &Node{
					Parent:    nodes[3],
					Id:        7,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(2)`,
					Typename:  "int",
					Value:     2,
				})
				AddNode(nodes[4], &Node{
					Parent:    nodes[4],
					Id:        10,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `int(3)`,
					Typename:  "int",
					Value:     3,
				})

				AddNode(GetNode(nodes[1], 0), nodes[2])
				AddNode(GetNode(nodes[1], 1), nodes[3])
				AddNode(GetNode(nodes[1], 2), nodes[4])

			})
		},
	}
}
func simple3ElementIntArray123() testData {
	value := [3]int{1, 2, 3}
	return testData{
		name:  "Simple 3-element int array: 1, 2, 3",
		value: value,
		want:  wantValue(`[3]int`, `[3]int{1,2,3,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.ArrayNode,
					Name:      "[3]int",
					Typename:  "[3]int",
					Value:     value,
				},
			}, func(nodes typegen.Nodes) {
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.ElementNode,
					Name:      `Index 0`,
					Typename:  "element",
					Value:     0,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     1,
					Id:        4,
					Type:      typegen.ElementNode,
					Name:      `Index 1`,
					Typename:  "element",
					Value:     1,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     2,
					Id:        6,
					Type:      typegen.ElementNode,
					Name:      `Index 2`,
					Typename:  "element",
					Value:     2,
				})

				AddNode(GetNode(nodes[1], 0), &Node{
					Parent:    GetNode(nodes[1], 0),
					Id:        3,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 0`,
					Typename:  "int",
					Value:     1,
				})
				AddNode(GetNode(nodes[1], 1), &Node{
					Parent:    GetNode(nodes[1], 1),
					Id:        5,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 1`,
					Typename:  "int",
					Value:     2,
				})
				AddNode(GetNode(nodes[1], 2), &Node{
					Parent:    GetNode(nodes[1], 2),
					Id:        7,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 2`,
					Typename:  "int",
					Value:     3,
				})
			})
		},
	}
}
func emptyStringIntMap() testData {
	value := map[string]int{}
	return testData{
		name:  "Empty string/int map",
		value: value,
		want:  wantValue("map[string]int", "map[string]int{}"),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.MapNode,
					Name:      "map[string]int",
					Typename:  "map[string]int",
					Value:     value,
				},
			}, nil)
		},
	}
}
func simple3ElementIntSlice123() testData {
	value := []int{1, 2, 3}
	return testData{
		name:  "Simple 3-element int slice: 1, 2, 3",
		value: value,
		want:  wantValue(`[]int`, `[]int{1,2,3,}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.SliceNode,
					Name:      "[]int",
					Typename:  "[]int",
					Value:     value,
				},
			}, func(nodes typegen.Nodes) {
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     0,
					Id:        2,
					Type:      typegen.ElementNode,
					Name:      `Index 0`,
					Typename:  "element",
					Value:     0,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     1,
					Id:        4,
					Type:      typegen.ElementNode,
					Name:      `Index 1`,
					Typename:  "element",
					Value:     1,
				})
				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Marshaler: m,
					Index:     2,
					Id:        6,
					Type:      typegen.ElementNode,
					Name:      `Index 2`,
					Typename:  "element",
					Value:     2,
				})

				AddNode(GetNode(nodes[1], 0), &Node{
					Parent:    GetNode(nodes[1], 0),
					Id:        3,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 0`,
					Typename:  "int",
					Value:     1,
				})
				AddNode(GetNode(nodes[1], 1), &Node{
					Parent:    GetNode(nodes[1], 1),
					Id:        5,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 1`,
					Typename:  "int",
					Value:     2,
				})
				AddNode(GetNode(nodes[1], 2), &Node{
					Parent:    GetNode(nodes[1], 2),
					Id:        7,
					Marshaler: m,
					Index:     0,
					Type:      typegen.IntNode,
					Name:      `Value 2`,
					Typename:  "int",
					Value:     3,
				})
			})
		},
	}
}
func emptyIntArray() testData {
	value := [0]int{}
	return testData{
		name:  "Empty array",
		value: value,
		want:  wantValue(`[0]int`, `[0]int{}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.ArrayNode,
					Name:      "[0]int",
					Typename:  "[0]int",
					Value:     value,
				},
			}, nil)
		},
	}
}
func simpleInterfaceContainingInt10() testData {
	value := interface{}(10)
	return testData{
		name:  "Simple interface containing int",
		value: value,
		want:  wantValue(`int`, `10`),
		nodes: func(m *nM) Nodes {
			return FixupNodes(Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.IntNode,
					Name:      `int(10)`,
					Typename:  "int",
					Value:     value,
				},
			}, nil)
		},
	}
}
func anySliceOfReflectValueOf10() testData {
	value := []any{reflect.ValueOf(10)}
	return testData{
		name:  "[]any{reflect.ValueOf(10)}",
		value: value,
		want:  wantValue(`[]any`, `[]any{reflect.ValueOf(10),}`),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.SliceNode,
					Name:      "[]interface {}",
					Typename:  "[]interface {}",
					Value:     value,
				},
				{
					Marshaler: m,
					Id:        3,
					Name:      "Value 0",
					Type:      typegen.InterfaceNode,
					Typename:  "any(reflect.Value)",
					Value:     reflect.ValueOf(10),
				},
			}, func(nodes typegen.Nodes) {

				AddNode(nodes[1], &Node{
					Parent:    nodes[1],
					Id:        2,
					Marshaler: m,
					Name:      "Index 0",
					Typename:  "element",
					Type:      typegen.ElementNode,
					Value:     0,
				})
				AddNode(GetNode(nodes[1], 0), nodes[2])
				nodes[2].Parent = GetNode(nodes[1], 0)

				AddNode(nodes[2], &Node{
					Parent:    nodes[2],
					Id:        4,
					Marshaler: m,
					Name:      "substitution",
					Typename:  "string",
					Type:      typegen.SubstitutionNode,
					Value:     `reflect.ValueOf(10)`,
				})

			})
		},
	}
}
func pointerToStructWithPropertyPointingToItself() testData {
	type recurStruct struct {
		name  string
		recur *recurStruct
		extra string
	}
	recur := recurStruct{name: "root", extra: "whatever"}
	recur.recur = &recur

	return testData{
		name:  "Pointer to struct with property pointing to itself",
		value: &recur,
		want:  wantPtrValue(`recurStruct`, `recurStruct{name:"root",recur:nil,extra:"whatever",}%s  var1.recur = &var1`, "\n"),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.PointerNode,
					Name:      "*typegen_test.recurStruct",
					Typename:  "*typegen_test.recurStruct",
					Value:     &recur,
				},
				{
					Marshaler: m,
					Id:        2,
					Name:      "typegen_test.recurStruct",
					Typename:  "typegen_test.recurStruct",
					Type:      typegen.StructNode,
					Value:     recur,
				},
			}, func(nodes typegen.Nodes) {

				AddNode(nodes[1], nodes[2])

				nodes[2].Parent = nodes[1]

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Parent:    nodes[2],
					Id:        3,
					Index:     0,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "name",
					Value:     nil,
				})

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Parent:    nodes[2],
					Id:        5,
					Index:     1,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "recur",
					Value:     nil,
				})

				nodes[1].Parent = GetNode(nodes[2], 1)

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Parent:    nodes[2],
					Id:        6,
					Index:     2,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "extra",
					Value:     nil,
				})

				AddNode(GetNode(nodes[2], 0), &Node{
					Marshaler: m,
					Parent:    GetNode(nodes[2], 0),
					Id:        4,
					Name:      `string("root")`,
					Typename:  "string",
					Type:      typegen.StringNode,
					Index:     0,
					Value:     "root",
				})
				AddNode(GetNode(nodes[2], 1), nodes[1])
				AddNode(GetNode(nodes[2], 2), &Node{
					Marshaler: m,
					Parent:    GetNode(nodes[2], 2),
					Id:        7,
					Name:      `string("whatever")`,
					Typename:  "string",
					Type:      typegen.StringNode,
					Value:     "whatever",
					Index:     0,
				})

			})
		},
	}
}
func pointerToStructWithIndirectPropertyPointingToItself() testData {
	type recurStruct struct {
		recur []*recurStruct
	}
	recur := recurStruct{}
	recur.recur = make([]*recurStruct, 1)
	recur.recur[0] = &recur

	return testData{
		name:  "Pointer to struct with indirect property pointing to itself",
		value: &recur,
		want:  wantPtrValue(`recurStruct`, `recurStruct{recur:nil,}%s  var2 := []*recurStruct{nil,}%s  var1.recur = var2%s  var2[0] = &var1`, "\n", "\n", "\n"),
		nodes: func(m *nM) typegen.Nodes {
			return FixupNodes(typegen.Nodes{
				nil,
				{
					Marshaler: m,
					Id:        1,
					Type:      typegen.PointerNode,
					Name:      "Value 0",
					Typename:  "*typegen_test.recurStruct",
					Value:     &recur,
				},
				{
					Marshaler: m,
					Id:        2,
					Name:      "typegen_test.recurStruct",
					Typename:  "typegen_test.recurStruct",
					Type:      typegen.StructNode,
					Value:     recur,
				},
				{
					Marshaler: m,
					Id:        4,
					Name:      "[]*typegen_test.recurStruct",
					Typename:  "[]*typegen_test.recurStruct",
					Type:      typegen.SliceNode,
					Value:     []any{"<example>"},
				},
			}, func(nodes typegen.Nodes) {

				nodes[2].Parent = nodes[1]
				AddNode(nodes[1], nodes[2])

				AddNode(nodes[2], &Node{
					Marshaler: m,
					Parent:    nodes[2],
					Id:        3,
					Index:     0,
					Typename:  "field",
					Type:      typegen.FieldNode,
					Name:      "recur",
				})

				AddNode(GetNode(nodes[2], 0), nodes[3])
				nodes[3].Parent = GetNode(nodes[2], 0)

				AddNode(nodes[3], &Node{
					Marshaler: m,
					Parent:    nodes[3],
					Id:        5,
					Name:      `Index 0`,
					Value:     0,
					Typename:  "element",
					Type:      typegen.ElementNode,
				})

				AddNode(GetNode(nodes[3], 0), nodes[1])
				nodes[1].Parent = GetNode(nodes[3], 0)

			})
		},
	}
}
