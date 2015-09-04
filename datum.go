package flowgraph

import (
	"fmt"
	"reflect"
	"strconv"
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

// EqualsTest returns true if two empty interfaces have the same numeric or string value.
func EqualsTest(n *Node, a,b Datum) bool {

	a2,b2,same := Promote(n, a, b)
	if !same { return false }

	switch a2.(type) {
	case string: { return a2.(string)==b2.(string) }		
        case int8: { return a2.(int8)==b2.(int8) }
        case uint8: { return a2.(uint8)==b2.(uint8) }
        case int16: { return a2.(int16)==b2.(int16) }
        case uint16: { return a2.(uint16)==b2.(uint16) }
        case int32: { return a2.(int32)==b2.(int32) }
        case uint32: { return a2.(uint32)==b2.(uint32) }
	case int64: { return a2.(int64)==b2.(int64) }
        case uint64: { return a2.(uint64)==b2.(uint64) }
	case int: { return a2.(int)==b2.(int) }
	case uint: { return a2.(uint)==b2.(uint) }
	case float32: { return a2.(float32)==b2.(float32) }
	case float64: { return a2.(float64)==b2.(float64) }
	case complex64: { return a2.(complex64)==b2.(complex64) }
	case complex128: { return a2.(complex128)==b2.(complex128) }
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

// String returns a string representation of a Datum with 
// ellipse shortened slices if TraceLevel<VVVV.
func String(d Datum) string {
       
	if IsSlice(d) {
		return StringSlice(d)
	}
        if dd,ok := d.([]Datum); ok {
		var s string
		for i := range dd {
			if i!= 0 { s+="|" }
			s += String(dd[i])
		}
		return s
	}
	if dd,ok := d.(nodeWrap); ok {
		return String(dd.datum)
	}
	if IsStruct(d) {
		return StringStruct(d)
	}
	switch d.(type) {
	case bool, int8, uint8, int16,uint16, int32, uint32, int64, uint64, int, uint: {
		return fmt.Sprintf("%v", d)
	}
	case float32, float64: {
		return fmt.Sprintf("%.4g", d)
	}
	case complex64, complex128: {
		return fmt.Sprintf("%.4g", d)
	}
	case string: {
		return fmt.Sprintf("%q", d)
	}
	}
	return fmt.Sprintf("%T(%+v)", d, d)
}

// StringSlice returns a string representation of a slice, ellipse shortened if TraceLevel<VVVV.
func StringSlice(d Datum) string {
	m := 8
	l := Len(d)
	if l < m || TraceLevel==VVVV { m = l }
	dv := reflect.ValueOf(d)
	dt := reflect.TypeOf(d)
	dts := dt.String()
	s := fmt.Sprintf("[:%d]%s([", dv.Len(), dts[2:len(dts)])
	for i := 0; i<m; i++ {
		if i!=0 {s += " "}
		s += fmt.Sprintf("%s", String(Index(d,i)))
	}
	if m<l && TraceLevel<VVVV {s += " ..."}
	s += "])"
	return s
}

// StringStruct returns a string representation of a struct with 
// ellipse shortened slices if TraceLevel<VVVV.
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

// ParseDatum parses a string for numeric constants, otherwise returns the string.
func ParseDatum(s string) Datum {
	var v Datum
	i32,err := strconv.ParseInt(s, 10, 32)
	if err==nil { v = int(i32); return v }
	i64,err := strconv.ParseInt(s, 10, 64)
	if err==nil { v = i64;  return v }
	f32,err := strconv.ParseFloat(s, 32)
	if err==nil { v = f32; return v }
	f64,err := strconv.ParseFloat(s, 64)
	if err==nil { v = f64; return v }
	if s=="true"||s=="false" { 
		b,err := strconv.ParseBool(s) 
		if err==nil { v = b; return v }
	}
	v = s
	return v
}
