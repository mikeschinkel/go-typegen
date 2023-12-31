package typegen_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mikeschinkel/go-diffator"
	"github.com/mikeschinkel/go-typegen"
	"github.com/stretchr/testify/assert"
)

type nM = typegen.NodeMarshaler
type nodesFunc func(m *nM) typegen.Nodes
type Node = typegen.Node
type Nodes = typegen.Nodes

var AddNode = typegen.TestAddNode
var GetNode = typegen.TestGetNode
var FixupNodes = typegen.TestFixupNodes

type testData struct {
	name      string
	value     any
	nodes     nodesFunc
	skipNodes bool
	want      string
}

func TestNodeBuilder_Marshal(t *testing.T) {
	tests := []testData{
		int100Node(),
		int64Node(),
		boolNode(),
		stringNode(),
		float64Node(),
		pointerToSimpleStructNode(),
		emptyIntSliceNode(),
		nilNode(),
		pointerToInterfaceStructContainingInterfacesNode(),
		simpleStringIntMapNode(),
		pointerToSimpleStruct(),
		sliceOfAnyContainingHelloGoodbye(),
		simpleAnySliceAllSameNumbers(),
		simpleAnySlice123(),
		simple3ElementIntArray123(),
		emptyStringIntMap(),
		simple3ElementIntSlice123(),
		emptyIntArray(),
		simpleInterfaceContainingInt10(),
		anySliceOfReflectValueOf10(),
		pointerToStructWithPropertyPointingToItself(),
		pointerToStructWithIndirectPropertyPointingToItself(),
	}
	subs := typegen.Substitutions{
		reflect.TypeOf(reflect.Value{}): func(rv *reflect.Value) string {
			return fmt.Sprintf("reflect.ValueOf(%v)", (*rv).Interface())
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
	nodeType := reflect.TypeOf((*typegen.NodeType)(nil)).Elem()
	comparator := diffator.NewObjectComparator(want, got, &diffator.ObjectOpts{
		PrettyPrint: diffator.Bool(true),
		FormatFunc: func(rt reflect.Type, a any) (s string) {
			switch {
			case rt == nil:
				s = "<invalid>"
			case rt == nodeType:
				s = typegen.NodeType(a.(uint64)).String()
			case rt.Kind() == reflect.UnsafePointer:
				s = "<unsafe-pointer>"
			default:
				s = fmt.Sprintf("%v", a)
			}
			return s
		},
	})
	return comparator.Compare()
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
