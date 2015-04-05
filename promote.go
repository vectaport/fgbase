package flowgraph

import (
	"reflect"
)

func biggerType(a, b Datum) bool {
	switch a.(type) {
	case bool: { }
	case int8, uint8: {
		switch b.(type) {
		case bool: {return true} 
		}
	}
	case int16, uint16: {
		switch b.(type) {
		case bool,int8,uint8: {return true} 
		}
	}
	case int32, uint32: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16: {return true} 
		}
	}
	case int, uint: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16,int32,uint32: {return true} 
		}
	}
	case int64, uint64: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16,int32,uint32,int,uint: {return true} 
		}
	}
	case float32: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16,int32,uint32,int64,uint64,int,uint: {return true} 
		}
	}
	case float64,complex64: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16,int32,uint32,int64,uint64,int,uint,float32: {return true} 
		}
	}
	case complex128: {
		switch b.(type) {
		case bool,int8,uint8,int16,uint16,int32,uint32,int64,uint64,int,uint,float32,float64,complex64: {return true} 
		}
	}
	}
	return false
}

// Promote pair of numeric empty interfaces (Datum) as necessary
func Promote(a, b Datum) (abig, bbig Datum, same bool) {

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)

	if(ta==tb) { return a,b,true }

	aBigger := biggerType(a, b)
	if (aBigger) {
		if(tb.ConvertibleTo(ta)) { return a,reflect.ValueOf(b).Convert(ta).Interface(),true }
		if(ta.ConvertibleTo(tb)) { return reflect.ValueOf(a).Convert(tb).Interface(),b,true }
	} else {
		if(ta.ConvertibleTo(tb)) { return reflect.ValueOf(a).Convert(tb).Interface(),b,true }
		if(tb.ConvertibleTo(ta)) { return a,reflect.ValueOf(b).Convert(ta).Interface(),true }
	}

	return a,b,false
}

