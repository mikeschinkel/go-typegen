package typegen_test

import (
	"fmt"
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

func TestNodeBuilder_Marshal(t *testing.T) {
	recur := recurStruct{name: "root", extra: "whatever"}
	recur.recur = &recur

	recur2 := recur2Struct{}
	recur2.recur = make([]*recur2Struct, 1)
	recur2.recur[0] = &recur2

	iFace := iFaceStruct{
		iFace1: interface{}("Hello"),
		iFace2: any(10),
	}

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "Pointer to interface struct containing interface{}(string) and any(int)",
			value: &iFace,
			want:  wantPtrValue(`iFaceStruct`, `iFaceStruct{iFace1:nil,iFace2:nil,}%s  var2 := any("Hello")%s  var3 := any(10)%s  var1.iFace1 = var2%s  var1.iFace2 = var3`, "\n", "\n", "\n", "\n"),
		},
		{
			name:  "nil",
			value: nil,
			want:  wantValue(`error`, `nil`),
		},
		{
			name:  "Simple string/int map",
			value: map[string]int{"Foo": 1, "Bar": 2, "Baz": 3},
			// Keys will be sorted alphabetically on output
			want: wantValue("map[string]int", `map[string]int{"Bar":2,"Baz":3,"Foo":1,}`),
		},
		{
			name:  "Empty string/int map",
			value: map[string]int{},
			want:  wantValue("map[string]int", "map[string]int{}"),
		},
		{
			name:  "Boolean true",
			value: true,
			want:  wantValue("bool", `true`),
		},
		{
			name:  "Integer",
			value: 100,
			want:  wantValue("int", `100`),
		},
		{
			name:  "64-bit integer",
			value: int64(100),
			want:  wantValue("int64", `int64(100)`),
		},
		{
			name:  "Float",
			value: 1.23,
			want:  wantValue("float64", `float64(1.230000)`),
		},
		{
			name:  "Simple String",
			value: "Hello World",
			want:  wantValue("string", `"Hello World"`),
		},
		{
			name:  "Empty int slice",
			value: []int{},
			want:  wantValue(`[]int`, `[]int{}`),
		},
		{
			name:  "Simple int slice",
			value: []int{1, 2, 3},
			want:  wantValue(`[]int`, `[]int{1,2,3,}`),
		},
		{
			name:  "Pointer to simple struct",
			value: &testStruct{},
			want:  wantPtrValue(`testStruct`, `testStruct{Int:0,String:"",}`),
		},
		{
			name:  "Pointer to struct with indirect property pointing to itself",
			value: &recur2,
			want:  wantPtrValue(`recur2Struct`, `recur2Struct{recur:nil,}%s  var2 := []*recur2Struct{&var1,}%s  var1.recur = var2`, "\n", "\n"),
		},
		{
			name:  "Pointer to struct with property pointing to itself",
			value: &recur,
			want:  wantPtrValue(`recurStruct`, `recurStruct{name:"root",recur:nil,extra:"whatever",}%s  var1.recur = &var1`, "\n"),
		},
		{
			name:  "Empty array",
			value: [0]int{},
			want:  wantValue(`[0]int`, `[0]int{}`),
		},
		{
			name:  "Simple int array",
			value: [3]int{1, 2, 3},
			want:  wantValue(`[3]int`, `[3]int{1,2,3,}`),
		},
		{
			name:  "Simple interface containing int",
			value: interface{}(10),
			want:  wantValue(`int`, `10`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := typegen.NewNodeMarshaler(tt.value)
			nodes := m.Marshal()
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
