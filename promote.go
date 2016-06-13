package flowgraph

import (
	"reflect"
)

func biggerType(a, b interface{}) bool {
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

// Promote pair of numeric empty interfaces (interface{}) as necessary.
func Promote(n *Node, a, b interface{}) (aBig, bBig interface{}, same bool) {

	var debug = false && n!=nil
	
	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if debug {
		n.Tracef("promote:  ta %v of Kind %s, tb %v of Kind %s\n", ta, ta.Kind().String(), tb, tb.Kind().String())
	}

	if ta==tb { return a,b,true }

	if ta==nil || tb==nil { return a,b,false }

	if ta.Kind()==tb.Kind() {
		if ta.Kind().String()==ta.String() {
			if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case -2: promoting %v to %v\n", tb, ta) }
			return a,reflect.ValueOf(b).Convert(ta).Interface(),true
		} 
		if tb.Kind().String()==tb.String() {
			if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case -1: promoting %v to %v\n", ta, tb) }
			return reflect.ValueOf(a).Convert(tb).Interface(),b,true
		}
	}

	aBigger := biggerType(a, b)
	if aBigger {
		if tb.ConvertibleTo(ta) { 
			if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case 0: promoting %v to %v\n", tb, ta) }
			return a,reflect.ValueOf(b).Convert(ta).Interface(),true 
		}
	}

	if ta.ConvertibleTo(tb) { 
		if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case 1: promoting %v to %v\n", ta, tb) }
		return reflect.ValueOf(a).Convert(tb).Interface(),b,true 
	}

	if !aBigger {
		if tb.ConvertibleTo(ta) { 
			if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case 2: promoting %v to %v\n", tb, ta) }
			return a,reflect.ValueOf(b).Convert(ta).Interface(),true 
		}
	}

	if debug && TraceLevel>=VVV && n!=nil { n.Tracef("case 3: no promotion between %v to %v\n", tb, ta) }
	return a,b,false
}

