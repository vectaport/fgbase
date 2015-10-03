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

	return a2==b2
}

// IsInt returns true if empty interface (Datum) is an int.
func IsInt (d Datum) bool {
	return reflect.ValueOf(d).Kind()==reflect.Int
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

	var s string
	switch d.(type) {
	case bool, int8, int16, int32, int64, int: {
		s = fmt.Sprintf("%v", d)
	}
	case uint8, uint16, uint32, uint64, uint: {
		w := 2
		switch d.(type) {
		case uint16: {
			w = 4
		}
		case uint32: {
			w = 8
		}
		case uint64, uint: {
			w = 16
		}
		}
		s = fmt.Sprintf("0x%0"+strconv.Itoa(w)+"x", d)
	}
	case float32, float64: {
		s = fmt.Sprintf("%.4g", d)
	}
	case complex64, complex128: {
		s = fmt.Sprintf("%.4g", d)
	}
	case string: {
		s = fmt.Sprintf("%q", d)
	}
	}

	
	if !TraceTypes && s!="" {
		return s

	}
	if s=="" {
		s = fmt.Sprintf("%v", d)
	}
	if IsInt(d) && !TraceTypes {
		return fmt.Sprintf("%s", s)
	} else {
		return fmt.Sprintf("%T(%s)", d, s)
	}
}

// StringSlice returns a string representation of a slice, ellipse shortened if TraceLevel<VVVV.
func StringSlice(d Datum) string {
	m := 8
	l := Len(d)
	if l < m || TraceLevel==VVVV { m = l }
	dv := reflect.ValueOf(d)
	dt := reflect.TypeOf(d)
	dts := dt.String()
	s := fmt.Sprintf("[:%d]%s{", dv.Len(), dts[2:len(dts)])
	for i := 0; i<m; i++ {
		if i!=0 {s += " "}
		s += fmt.Sprintf("%s", String(Index(d,i)))
	}
	if m<l && TraceLevel<VVVV {s += " ..."}
	s += "}"
	return s
}

// isShadowSlice returns true if the nth field of the struct is a slice that is shadowing
// another field.
func isShadowSlice(d Datum, nth int) bool {
	if !IsStruct(d) {
		return false
	}

	dv := reflect.ValueOf(d)
	l := dv.NumField()
	if nth>=l {
		return false
	}
	shadow := dv.Type().Field(nth)

	// search for shadowing struct
	for j :=0; j<l; j++ {
		if j==nth { continue }
		slice := dv.Type().Field(j)
		if shadow.Type.String()== slice.Type.String() && shadow.Name=="Shadow"+slice.Name {
			return true
		}
	}

	return false
}


// shadowSlice returns the index of a struct field that is a slice that
// is being shadowed by the nth field of the struct (with a "Shadow" prefix and matching type).  
// -1 returned if not found
func shadowSlice(d Datum, nth int) int {
	if !IsStruct(d) {
		return -1
	}

	dv := reflect.ValueOf(d)
	l := dv.NumField()
	if nth>=l {
		return -1
	}
	slice := dv.Type().Field(nth)

	// search for shadowed struct
	for j :=0; j<l; j++ {
		if j==nth { continue }
		shadow := dv.Type().Field(j)
		if shadow.Type.String()== slice.Type.String() && shadow.Name=="Shadow"+slice.Name {
			return j
		}
	}

	return -1
}


// shadowString returns a struct-like string for a shadows slice, where the index of a changed
// value proceeds the value and a colon.
func shadowString(d Datum, sliceIndex, shadowIndex int) string {
	if !IsStruct(d) {
		return ""
	}

	dv := reflect.ValueOf(d)
	n := dv.NumField()
	if sliceIndex>=n || shadowIndex>=n {
		return ""
	}

	slice := dv.Field(sliceIndex).Interface()
	shadow := dv.Field(shadowIndex).Interface()
	st := reflect.TypeOf(slice).String()

	l := Len(slice)
	s := fmt.Sprintf("[:%d]%v{", l, st[2:])

	var first = true
	for i := 0; i<l; i++ {
		if !EqualsTest(nil, Index(slice, i), Index(shadow, i)) {
			if !first {
				s += " "
			} else {
				first = false
			}
			s += fmt.Sprintf("%d:%s", i, String(Index(slice, i)))
		}
	}
	s += "}"
	return s
}


// StringStruct returns a string representation of a struct with 
// ellipse shortened slices if TraceLevel<VVVV.
func StringStruct(d Datum) string {
	dv := reflect.ValueOf(d)
	l := dv.NumField()
	var s string
	if TraceLevel >= VVVV {
		s = fmt.Sprintf("%T", d)
	}
	s += "{"
	flg := false
	for i := 0; i<l; i++ {
		ft := dv.Type().Field(i)
		if ft.Name[0]>='A' && ft.Name[0]<='Z' {
			if isShadowSlice(d, i) {
				continue
			}
			if flg { 
				s += " " 
			} else {
				flg = true
			}
			s += ft.Name
			s += ":"
			j := shadowSlice(d, i)
			if j<0 {
				s += String(dv.Field(i).Interface())
			} else {
				s += shadowString(d, i, j)
			}
		}
	}
	s += "}"
	return s
}

// ParseDatum parses a string for numeric constants, otherwise returns the string.
func ParseDatum(s string) Datum {
	var v Datum

	// trim trailing whitespace or comments
	var s2 string
	for i := range s {
		if s[i]=='#' { break }
		if len(s)>i+1 && s[i:i+2]=="//" { break }
		if s[i]==' ' || s[i]=='\t' { break }
		s2 += s[i:i+1]
	}
	s = s2

	if s=="{}" {
		return Nada{}
	}
	
	if len(s)>2 && s[0:2]=="0x" {
		s = s[2:]
		if len(s)<=2 {
			u8,err := strconv.ParseUint(s, 16, 8)
			if err==nil { v = uint8(u8); return v }
		}
		if len(s)<=4 {
			u16,err := strconv.ParseUint(s, 16, 16)
			if err==nil { v = uint16(u16); return v }
		}
		if len(s)<=8 {
			u32,err := strconv.ParseUint(s, 16, 32)
			if err==nil { v = uint32(u32); return v }
		}
		u64,err := strconv.ParseUint(s, 16, 64)
		if err==nil { v = u64;  return v }
	}
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
