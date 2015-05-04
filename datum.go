package flowgraph

import (
	"fmt"
	"reflect"
)

// Datum is an empty interface for generic data flow.
type Datum interface{}

// ZeroTest returns true if empty interface (Datum) is a numeric zero.
func ZeroTest(a Datum) bool {

	switch a.(type) {
        case int8: { return a.(int8)==0 }
        case uint8: { return a.(uint8)==0 }
        case int16: { return a.(int16)==0 }
        case uint16: { return a.(uint16)==0 }
        case int32: { return a.(int32)==0 }
        case uint32: { return a.(uint32)==0 }
	case int64: { return a.(int64)==0 }
        case uint64: { return a.(uint64)==0 }
	case int: { return a.(int)==0 }
	case uint: { return a.(uint)==0 }
	case float32: { return a.(float32)==0.0 }
	case float64: { return a.(float64)==0.0 }
	case complex64: { return a.(complex64)==0.0+0.0i }
	case complex128: { return a.(complex128)==0.0+0.0i }
	default: { return false }
	}
}

// IsSlice returns true if empty interface (Datum) is a slice.
func IsSlice (d Datum) bool {
	return reflect.ValueOf(d).Kind()==reflect.Slice
}

// IsStruct returns true if empty interface (Datum) is a slice.
func IsStruct (d Datum) bool {
	return reflect.ValueOf(d).Kind()==reflect.Struct
}

// Index returns the nth element of an empty interface (Datum) that is a slice.
func Index(d Datum, i int) Datum {
	return reflect.ValueOf(d).Index(i).Interface()
}

// Len returns the length of an empty interface (Datum) if it is a slice.
func Len(d Datum) int {
	if IsSlice(d) { 
		return reflect.ValueOf(d).Len()
	}
	return 0
}

// CopySlice returns a copy of a slice from an empty interface (as an empty interface).
func CopySlice(d Datum) Datum {
	dt := reflect.TypeOf(d)
	dv := reflect.ValueOf(d)
	r := reflect.MakeSlice(dt, dv.Len(), dv.Cap()).Interface()
	reflect.Copy(reflect.ValueOf(r), reflect.ValueOf(d))
	return r
}

// String returns a string representation of a Datum
func String(d Datum) string {

	if IsSlice(d) {
		return StringSlice(d)
	}
	if IsStruct(d) {
		return StringStruct(d)
	}
	switch d.(type) {
	case int8, uint8, int16, uint16, int32, uint32, int64, uint64,
		int, uint, float32, float64, complex64, complex128, string: {
			
			return fmt.Sprintf("%v", d)
		}
	default: {
		return fmt.Sprintf("%T(%v)", d)
	}
	}
}

// StringSlice returns a ellipse shortened string representation of a 
// slice when TraceLevel<VVVV.
func StringSlice(d Datum) string {
	if false {
		return fmt.Sprintf("%p[0:%d]", d, Len(d.(Interface2)))
	}
	m := 8
	l := Len(d)
	if l < m || TraceLevel==VVVV { m = l }
	s := fmt.Sprintf("%T([", d)
	for i := 0; i<m; i++ {
		if i!=0 {s += " "}
		s += fmt.Sprintf("%s", String(Index(d,i)))
	}
	if m<l && TraceLevel<VVVV {s += " ..."}
	s += "])"
	return s
}

// StringStruct returns a string representation of a struct with an
// ellipse shortened string representation of any slice when TraceLevel<VVVV.
func StringStruct(d Datum) string {
	dv := reflect.ValueOf(d)
	l := dv.NumField()
	s := fmt.Sprintf("%T({", d)
	flg := false
	for i := 0; i<l; i++ {
		ft := dv.Type().Field(i)
		if ft.Name[0]>='A' && ft.Name[0]<='Z' {
			if flg { 
				s += " " 
			} else {
				flg = true
			}
			s += ft.Name
			s += ":"
			s += String(dv.Field(i).Interface())
		}
	}
	s += "})"
	return s
}


