package typegen_test

import (
	"fmt"
	"strings"
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

func TestCodeBuilder_Marshal(t *testing.T) {
	recur := recurStruct{name: "root", extra: "whatever"}
	recur.recur = &recur

	recur2 := recur2Struct{}
	recur2.recur = make([]*recur2Struct, 1)
	recur2.recur[0] = &recur2

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "Pointer to struct with property pointing to itself",
			value: &recur,
			want:  ``,
		},
		//{
		//	name:  "Empty string/int map",
		//	value: map[string]int{},
		//	want:  `map[string]int{}`,
		//},
		//{
		//	name:  "Simple string/int map",
		//	value: map[string]int{"Foo": 1, "Bar": 2, "Baz": 3},
		//	want:  `map[string]int{"Bar":2,"Baz":3,"Foo":1,}`, // Keys sorted alphabetically
		//},
		//{
		//	name:  "Boolean true",
		//	value: true,
		//	want:  `true`,
		//},
		//{
		//	name:  "Integer",
		//	value: 100,
		//	want:  `100`,
		//},
		//{
		//	name:  "Float",
		//	value: 1.23,
		//	want:  `1.230000`,
		//},
		//{
		//	name:  "Simple String",
		//	value: "Hello World",
		//	want:  `"Hello World"`,
		//},
		//{
		//	name:  "Pointer to simple struct",
		//	value: &testStruct{},
		//	want:  `&typegen_test.testStruct{Int:0,String:"",}`,
		//},
		//{
		//	name:  "Simple struct",
		//	value: testStruct{},
		//	want:  `typegen_test.testStruct{Int:0,String:"",}`,
		//},
		//{
		//	name:  "Empty int slice",
		//	value: []int{},
		//	want:  `[]int{}`,
		//},
		//{
		//	name:  "Simple int slice",
		//	value: []int{1, 2, 3},
		//	want:  `[]int{1,2,3,}`,
		//},
		//{
		//	name:  "Indirect Pointer to struct with property pointing to itself",
		//	value: &recur2,
		//	want:  `&typegen.recur2Struct{recur:[]*typegen.recur2Struct{&typegen.recur2Struct{recur:[]*typegen.recur2Struct{nil/*** recursion ***/,},},},}`,
		//},
		//{
		//	value: struct{}{},
		//	want:  `struct {}{}`,
		//},
		//{
		//	name:  "Struct with property pointing to itself",
		//	value: recur,
		//	want:  `typegen.recurStruct{name:"root",recur:&typegen.recurStruct{name:"root",recur:&typegen.recurStruct{name:"root",recur:nil/*** recursion ***/,extra:"whatever",},extra:"whatever",},extra:"whatever",}`,
		//},
		//{
		//	name: "interface{}{}",
		//	//value: Effectively as if interface{}{}
		//	want: `nil`,
		//},
		//{
		//	name:  "nil",
		//	value: nil,
		//	want:  `nil`,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := typegen.NewCodeBuilder(tt.value, "getData")
			cb.Build()
			got := cb.String()
			//g := typegen.NewGenerator()
			//g.WriteCode(cb.Nodes()[0])
			//got := g.String()
			if strings.HasPrefix(tt.want, "map[") {
				assert.Equal(t, tt.want, fmt.Sprint(got))
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

//// Start with the type definition
//case reflect.Struct:
//case reflect.Ptr:

// 824633795064