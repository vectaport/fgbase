package flowgraph

import (
	"reflect"
)

// promote pair of numeric values as necessary
func Promote(a, b Datum) (abig, bbig Datum) {

	if(reflect.TypeOf(a)==reflect.TypeOf(b)) {
		return a,b
	}
	
	switch a.(type) {
	case bool: {
		aa := a.(bool);
		switch b.(type) {
		case int8: { bb := b.(int8); if (aa) { return int8(1),bb } else { return int8(0),bb }}
                case uint8: { bb := b.(uint8); if (aa) { return uint8(1),bb } else { return uint8(0),bb }}
		case int16: { bb := b.(int16);  if (aa) { return int16(1),bb } else { return int16(0),bb }}
                case uint16: { bb := b.(uint16); if (aa) { return uint16(1),bb } else { return uint16(0),bb }}
                case int32: { bb := b.(int32); if (aa) { return int32(1),bb } else { return int32(0),bb } }
                case uint32: { bb := b.(uint32); if (aa) { return uint32(1),bb } else { return uint32(0),bb }}
                case int64: { bb := b.(int64); if (aa) { return int64(1),bb } else { return int64(0),bb } }
                case uint64: { bb := b.(uint64); if (aa) { return uint64(1),bb } else { return uint64(0),bb }}
                case int: { bb := b.(int); if (aa) { return int(1),bb } else { return int(0),bb }}
                case uint: { bb := b.(uint); if (aa) { return uint(1),bb } else { return uint(0),bb }}
	        case float32: { bb := b.(float32); if (aa) { return float32(1.),bb } else { return float32(0.),bb }}
	        case float64: { bb := b.(float64); if (aa) { return float64(1.),bb } else { return float64(0.),bb }}
	        case complex64: { bb := b.(complex64); if (aa) { return complex64(float32(1.)),bb } else { return complex64(float32(0.)),bb }}
	        case complex128: { bb := b.(complex128); if (aa) { return complex128(float64(1.)),bb } else { return complex128(float64(0.)),bb }}
		}
	}
	case int8: {
		aa := a.(int8);
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,int8(1) } else { return aa,int8(0) }}
                case uint8: { bb := b.(uint8); return int16(aa),int16(bb) }
		case int16: { bb := b.(int16); return int16(aa),bb }
		case uint16: { bb := b.(uint16); return int32(aa),int32(bb) }
		case int32: { bb := b.(int32); return int32(aa),bb }
		case uint32: { bb := b.(uint32); return int64(aa),int64(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return aa,bb }  // no promotion
		case int: { bb := b.(int); return int(aa),bb }
		case uint: { bb := b.(uint); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case uint8: {
		aa := a.(uint8);
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,uint8(1) } else { return aa,uint8(0) }}
                case int8: { bb := b.(int8); return int16(aa),int16(bb) }
		case int16: { bb := b.(int16); return int16(aa),bb }
		case uint16: { bb := b.(uint16); return uint16(aa),bb }
		case int32: { bb := b.(int32); return int32(aa),bb }
		case uint32: { bb := b.(uint32); return uint32(aa),bb }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return uint64(aa),bb }
		case int: { bb := b.(int); return int(aa),bb }
		case uint: { bb := b.(uint); return int(aa),bb }
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case int16: {
		aa := a.(int16);
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,int16(1) } else { return aa,int16(0) }}
		case int8: { bb := b.(int8); return aa,int16(bb) }
		case uint8: { bb := b.(uint8); return aa,int16(bb) }
		case uint16: { bb := b.(uint16); return int32(aa),int32(bb) }
		case int32: { bb := b.(int32); return int32(aa),bb }
		case uint32: { bb := b.(uint32); return int64(aa),int64(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return aa,bb } // no promotion
		case int: { bb := b.(int); return int(aa),bb }
		case uint: { bb := b.(uint); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case uint16: {
		aa := a.(uint16);
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,uint16(1) } else { return aa,uint16(0) }}
		case int8: { bb := b.(int8); return int32(aa),int32(bb) }
		case uint8: { bb := b.(uint8); return aa,uint16(bb) }
		case int16: { bb := b.(int16); return int32(aa),int32(bb) }
		case int32: { bb := b.(int32); return int32(aa),bb }
		case uint32: { bb := b.(uint32); return uint32(aa),uint32(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return uint64(aa),bb }
		case int: { bb := b.(int); return int(aa),bb }
		case uint: { bb := b.(uint); return uint(aa),bb }
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case int32: {
		aa := a.(int32)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,int32(1) } else { return aa,int32(0) }}
		case int8: { bb := b.(int8); return aa,int32(bb) }
		case uint8: { bb := b.(uint8); return aa,int32(bb) }
		case int16: { bb := b.(int16); return aa,int32(bb) }
		case uint16: { bb := b.(uint16); return aa,int32(bb) }
		case uint32: { bb := b.(uint32); return int64(aa),int64(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return aa, bb } // no promotion
		case int: { bb := b.(int); return int(aa),bb }
		case uint: { bb := b.(uint); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case uint32: {
		aa := a.(uint32)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,uint32(1) } else { return aa,uint32(0) }}
		case int8: { bb := b.(int8); return int64(aa),int64(bb) }
		case uint8: { bb := b.(uint8); return aa,uint32(bb) }
		case int16: { bb := b.(int16); return int64(aa),int64(bb) }
		case uint16: { bb := b.(uint16); return aa,uint32(bb) }
		case int32: { bb := b.(int32); return int64(aa),int64(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return aa,uint64(bb) }
		case int: { bb := b.(int); return int64(aa),int64(bb) }
		case uint: { bb := b.(uint); return uint(aa),bb }
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case int64: {
		aa := a.(int64)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,int64(1) } else { return aa,int64(0) }}
		case int8: { bb := b.(int8); return aa,int64(bb) }
		case uint8: { bb := b.(uint8); return aa,int64(bb) }
		case int16: { bb := b.(int16); return aa,int64(bb) }
		case uint16: { bb := b.(uint16); return aa,int64(bb) }
		case int32: { bb := b.(int32); return aa,int64(bb) }
		case uint32: { bb := b.(uint32); return aa,int64(bb) }
		case uint64: { bb := b.(uint64); return aa,bb } // no promotion
		case int: { bb := b.(int); return aa,int64(bb) }
		case uint: { bb := b.(uint); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case uint64: {
		aa := a.(uint64)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,uint64(1) } else { return aa,uint64(0) }}
		case int8: { bb := b.(int8); return aa,bb } // no promotion
		case uint8: { bb := b.(uint8); return aa,uint64(bb) }
		case int16: { bb := b.(int16); return aa,bb } // no promotion
		case uint16: { bb := b.(uint16); return aa,uint64(bb) }
		case int32: { bb := b.(int32); return aa,bb } // no promotion
		case uint32: { bb := b.(uint32); return aa,uint64(bb) }
		case int64: { bb := b.(int64); return aa,bb } // no promotion
		case int: { bb := b.(int); return aa,bb } // no promotion
		case uint: { bb := b.(uint); return aa,uint64(bb) }
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case int: {
		aa := a.(int)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,int(1) } else { return aa,int(0) }}
		case int8: { bb := b.(int8); return aa,int(bb) }
		case uint8: { bb := b.(uint8); return aa,int(bb) }
		case int16: { bb := b.(int16); return aa,int(bb) }
		case uint16: { bb := b.(uint16); return aa,int(bb) }
		case int32: { bb := b.(int32); return aa,int(bb) }
		case uint32: { bb := b.(uint32); return int64(aa),int64(bb) }
		case int64: { bb := b.(int64); return int64(aa),bb }
		case uint64: { bb := b.(uint64); return aa,bb } // no promotion
		case uint: { bb := b.(uint); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case uint: {
		aa := a.(uint)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,uint(1) } else { return aa,uint(0) }}
		case int8: { bb := b.(int8); return aa,bb } // no promotion
		case uint8: { bb := b.(uint8); return aa,uint(bb) }
		case int16: { bb := b.(int16); return aa,bb } // no promotion
		case uint16: { bb := b.(uint16); return aa,uint(bb) }
		case int32: { bb := b.(int32); return aa,bb } // no promotion
		case uint32: { bb := b.(uint32); return aa,uint(bb) }
		case int64: { bb := b.(int64); return aa,bb } // no promotion
		case uint64: { bb := b.(uint64); return uint64(aa),bb }
		case int: { bb := b.(int); return aa,bb } // no promotion
	        case float32: { bb := b.(float32); return float32(aa),bb }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := float32(aa) + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case float32: {
		aa := a.(float32)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,float32(1.) } else { return aa,float32(0.) }}
		case int8: { bb := b.(int8); return aa,float32(bb) }
		case uint8: { bb := b.(uint8); return aa,float32(bb) }
		case int16: { bb := b.(int16); return aa,float32(bb) }
		case uint16: { bb := b.(uint16); return aa,float32(bb) }
		case int32: { bb := b.(uint32); return aa,float32(bb) }
		case uint32: { bb := b.(uint32); return aa,float32(bb) }
		case int64: { bb := b.(int64); return aa,float32(bb) }
		case uint64: { bb := b.(uint64); return aa,float32(bb) }
		case int: { bb := b.(int); return aa,float32(bb) }
		case uint: { bb := b.(uint); return aa,float32(bb) }
	        case float64: { bb := b.(float64); return float64(aa),bb }
	        case complex64: { bb := b.(complex64); aaa := aa + 0i; return aaa,bb }
	        case complex128: { bb := b.(complex128); aaa := float64(aa) + 0i; return aaa,bb }
		}
	}
	case float64: {
		aa := a.(float64)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { return aa,float64(1.) } else { return aa,float64(0.) }}
		case int8: { bb := b.(int8); return aa,float64(bb) }
		case uint8: { bb := b.(uint8); return aa,float64(bb) }
		case int16: { bb := b.(int16); return aa,float64(bb) }
		case uint16: { bb := b.(uint16); return aa,float64(bb) }
		case int32: { bb := b.(uint32); return aa,float64(bb) }
		case uint32: { bb := b.(uint32); return aa,float64(bb) }
		case int64: { bb := b.(int64); return aa,float64(bb) }
		case uint64: { bb := b.(uint64); return aa,float64(bb) }
		case int: { bb := b.(int); return aa,float64(bb) }
		case uint: { bb := b.(uint); return aa,float64(bb) }
	        case float32: { bb := b.(float64); return aa,float64(bb) }
	        case complex64: { bb := b.(complex64); aaa := aa + 0i; return aaa,complex128(bb) }
	        case complex128: { bb := b.(complex128); aaa := aa + 0i; return aaa,bb }
		}
	}
	case complex64: {
		aa := a.(complex64)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) { bbb := complex(float32(1.),0); return aa,bbb } else { bbb := complex(float32(0.),0); return aa,bbb }}
		case int8: { bb := b.(int8); bbb := complex(float32(bb),0); return aa,bbb }
		case uint8: { bb := b.(uint8); bbb := complex(float32(bb),0); return aa,bbb }
		case int16: { bb := b.(int16); bbb := complex(float32(bb),0); return aa,bbb }
		case uint16: { bb := b.(uint16); bbb := complex(float32(bb),0); return aa,bbb }
		case int32: { bb := b.(uint32);  bbb := complex(float32(bb),0); return aa,bbb }
		case uint32: { bb := b.(uint32);  bbb := complex(float32(bb),0); return aa,bbb }
		case int64: { bb := b.(int64);  bbb := complex(float32(bb),0); return aa,bbb }
		case uint64: { bb := b.(uint64);  bbb := complex(float32(bb),0); return aa,bbb }
		case int: { bb := b.(int);  bbb := complex(float32(bb),0); return aa,bbb }
		case uint: { bb := b.(uint);  bbb := complex(float32(bb),0); return aa,bbb }
	        case float32: { bb := b.(float32); bbb := complex(float32(bb),0); return aa,bbb }
	        case float64: { bb := b.(float64); bbb := complex(float64(bb),0); return complex128(aa),bbb }
	        case complex128: { bb := b.(complex128); return complex128(aa),bb }
		}
	}
	case complex128: {
		aa := a.(complex128)
		switch b.(type) {
		case bool: { bb := b.(bool); if (bb) {  bbb := complex(float64(1.),0); return aa,bbb } else { bbb := complex(float64(0.),0); return aa,bbb }}
		case int8: { bb := b.(int8); bbb := complex(float64(bb),0); return aa,bbb }
		case uint8: { bb := b.(uint8); bbb := complex(float64(bb),0); return aa,bbb }
		case int16: { bb := b.(int16); bbb := complex(float64(bb),0); return aa,bbb }
		case uint16: { bb := b.(uint16); bbb := complex(float64(bb),0); return aa,bbb }
		case int32: { bb := b.(uint32);  bbb := complex(float64(bb),0); return aa,bbb }
		case uint32: { bb := b.(uint32);  bbb := complex(float64(bb), 0); return aa,bbb }
		case int64: { bb := b.(int64);  bbb := complex(float64(bb),0); return aa,bbb }
		case uint64: { bb := b.(uint64);  bbb := complex(float64(bb),0); return aa,bbb }
		case int: { bb := b.(int);  bbb := complex(float64(bb),0); return aa,bbb }
		case uint: { bb := b.(uint);  bbb := complex(float64(bb),0); return aa,bbb }
	        case float32: { bb := b.(float32); bbb := complex(float64(bb),0); return aa,bbb }
	        case float64: { bb := b.(float64); bbb := complex(float64(bb),0); return complex128(aa),bbb }
	        case complex64: { bb := b.(complex128); return aa,complex128(bb) }
		}
	}
	}
	
	return a,b
}
