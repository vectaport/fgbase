package flowgraph

import (
	"reflect"
)

func rshFire2(a, b Datum) Datum {
	
	switch a.(type) {
        case uint8: { return a.(uint8)>>b.(uint8) }
        case uint16: { return a.(uint16)>>b.(uint16) }
        case uint32: { return a.(uint32)>>b.(uint32) }
        case uint64: { return a.(uint64)>>b.(uint64) }
	case uint: { return a.(uint)>>b.(uint) }
	default: { return nil }
	}
}

// Right shift primitive
func rshFire(n *Node) {

	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]

	aTmp,bTmp,same := Promote(n, a.Val, b.Val)

	if(!same) {
		n.Errorf("incompatible types for right shift (%v>>%v)", reflect.TypeOf(a.Val), reflect.TypeOf(b.Val))
		x.Val = nil
	} else {
		x.Val = rshFire2(aTmp, bTmp)
	}
}

// FuncRsh right shifts a value(x = a >> b).
func FuncRsh(a, b, x Edge) Node {

	node := MakeNode("rsh", []*Edge{&a, &b}, []*Edge{&x}, nil, rshFire)
	return node
}

