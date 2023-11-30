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
			name:      "[]any{reflect.ValueOf(10)}",
			value:     []any{reflect.ValueOf(10)},
			want:      wantValue(`[]any`, `[]any{reflect.ValueOf(10),}`),
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
		sliceOfAnyContainingHelloGoodbye(),
		simpleAnySliceAllSameNumbers(),
		simpleAnySlice123(),
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
				diff := getDiff(want, got)
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

func getDiff(want, got any) (diff string) {
	d := diffator.New()
	d.Pretty = true
	nodeType := reflect.TypeOf((*typegen.NodeType)(nil)).Elem()
	d.FormatFunc = func(rt diffator.ReflectTyper, a any) (s string) {
		switch {
		case rt.ReflectType() == nodeType:
			s = typegen.NodeType(a.(uint64)).String()
		default:
			s = fmt.Sprintf("%v", a)
		}
		return s
	}
	diff = d.Diff(want, got)
	return diff
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
