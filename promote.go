package flowgraph

import (
	"reflect"
)

// Promote pair of numeric values as necessary
func Promote(a, b Datum) (abig, bbig Datum) {

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)

	if(ta==tb) { return a,b }

	if (tb.ConvertibleTo(ta)) { return a,reflect.ValueOf(b).Convert(ta).Interface() }
	if (ta.ConvertibleTo(tb)) { return reflect.ValueOf(a).Convert(tb).Interface(),b }

	return a,b
}

