package typegen

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const RecursionDepth = 2

type refVal = reflect.Value
type reflectValueStack = Stack[[]reflect.Value, reflect.Value]

type CodeBuilder struct {
	value    any
	Prettify bool
	strings.Builder
	visited   reflectValueStack
	recursing reflectValueStack
}

func NewCodeBuilder() *CodeBuilder {
	cb := &CodeBuilder{}
	return cb
}

func (cb *CodeBuilder) Marshal(value any) (code string) {
	cb.value = value
	rv := reflect.ValueOf(value)

	// If the value is a pointer to an interface{}, we want to get the interface.
	// Then we can use .Elem() to get to the concrete value that the interface holds.
	switch rv.Kind() {
	case reflect.Interface:
		rv = rv.Elem() // Just dereference the interface.
	case reflect.Ptr:
		if rv.Elem().Kind() == reflect.Interface {
			rv = rv.Elem().Elem() // Twice .Elem(): one to dereference the pointer, one to get the interface value.
		}
	}
	cb.marshalValue(rv)
	return cb.String()
}

func (cb *CodeBuilder) checkRecursion(rv refVal) {
	var tryRV reflect.Value
	switch {
	case rv.Kind() == reflect.Pointer && rv.Elem().Kind() == reflect.Interface:
		tryRV = rv.Elem().Elem()
	case rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer:
		tryRV = rv.Elem() // Just dereference the interface.
	default:
		tryRV = rv
	}
	if cb.visited.Has(tryRV) {
		cb.recursing.Push(tryRV)
		goto end
	}
end:
	cb.visited.Push(tryRV)
}
func (cb *CodeBuilder) marshalField(fld field) {
	cb.WriteString(fld.name)
	cb.WriteByte(':')

	if fld.value.Kind() != reflect.Pointer {
		cb.marshalValue(fld.value)
		goto end
	}

	if cb.recursing.Depth() >= RecursionDepth {
		cb.WriteString("nil/*** recursion ***/")
		cb.recursing.Drop()
		goto end
	}

	cb.marshalValue(fld.value)
end:
}

func (cb *CodeBuilder) marshalElement(ele reflect.Value) {
	if ele.Kind() != reflect.Pointer {
		cb.marshalValue(ele)
		goto end
	}

	if cb.recursing.Depth() >= RecursionDepth {
		cb.WriteString("nil/*** recursion ***/")
		cb.recursing.Drop()
		goto end
	}

	cb.marshalValue(ele)
end:
}

func (cb *CodeBuilder) marshalValue(rv refVal) {

	cb.checkRecursion(rv)
	defer cb.visited.Drop()

	// Start with the type definition
	switch rv.Kind() {
	case reflect.Struct:
		cb.marshalStruct(rv)
	case reflect.Slice:
		cb.marshalSlice(rv)
	case reflect.Map:
		cb.marshalMap(rv)
	case reflect.Ptr:
		cb.marshalPtr(rv)
	case reflect.Interface:
		cb.marshalInterface(rv)

	case reflect.String:
		cb.WriteString(strconv.Quote(rv.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cb.WriteString(fmt.Sprintf("%d", rv.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cb.WriteString(fmt.Sprintf("%d", rv.Uint()))
	case reflect.Float32, reflect.Float64:
		cb.WriteString(fmt.Sprintf("%f", rv.Float()))
	case reflect.Bool:
		cb.WriteString(fmt.Sprintf("%t", rv.Bool()))
	case reflect.Invalid:
		cb.WriteString("nil")
	default:
		panicf("Unexpected type/kind: %s", rv.Kind().String())
	}
	//end:
}

func (cb *CodeBuilder) marshalSlice(rv refVal) {
	cb.WriteString(fmt.Sprintf("[]%s{", rv.Type().Elem()))
	for i := 0; i < rv.Len(); i++ {
		cb.marshalElement(rv.Index(i))
		cb.WriteByte(',')
	}
	cb.WriteString("}")
}

func (cb *CodeBuilder) marshalMap(rv refVal) {
	cb.WriteString(fmt.Sprintf("map[%s]%s{", rv.Type().Key(), rv.Type().Elem()))
	keyValues := rv.MapKeys()
	keys := make([]reflect.Value, len(keyValues))
	for i, k := range keyValues {
		keys[i] = k
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	for _, key := range keys {
		cb.marshalValue(key)
		cb.WriteByte(':')
		cb.marshalElement(rv.MapIndex(key))
		cb.WriteByte(',')
	}
	cb.WriteString("}")
}

func (cb *CodeBuilder) marshalPtr(rv refVal) {
	if rv.IsNil() {
		cb.WriteString("nil")
		return
	}
	cb.WriteByte('&')
	cb.marshalValue(rv.Elem())
}

func (cb *CodeBuilder) marshalInterface(rv refVal) {
	if rv.IsNil() {
		cb.WriteString("nil")
		return
	}
	cb.marshalValue(rv.Elem())
}

func (cb *CodeBuilder) marshalStruct(rv refVal) {
	cb.WriteString(fmt.Sprintf("%s{", rv.Type()))
	for i := 0; i < rv.NumField(); i++ {
		cb.marshalField(field{
			value: rv.Field(i),
			name:  rv.Type().Field(i).Name,
		})
		cb.WriteByte(',')
	}
	cb.WriteByte('}')
}

type field struct {
	value reflect.Value
	name  string
}
