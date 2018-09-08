package fgbase

import (
	"reflect"
)

func modFire2(a, b interface{}) interface{} {

	switch a.(type) {
	case int8:
		{
			return a.(int8) % b.(int8)
		}
	case uint8:
		{
			return a.(uint8) % b.(uint8)
		}
	case int16:
		{
			return a.(int16) % b.(int16)
		}
	case uint16:
		{
			return a.(uint16) % b.(uint16)
		}
	case int32:
		{
			return a.(int32) % b.(int32)
		}
	case uint32:
		{
			return a.(uint32) % b.(uint32)
		}
	case int64:
		{
			return a.(int64) % b.(int64)
		}
	case uint64:
		{
			return a.(uint64) % b.(uint64)
		}
	case int:
		{
			return a.(int) % b.(int)
		}
	case uint:
		{
			return a.(uint) % b.(uint)
		}
	default:
		{
			return nil
		}
	}
}

func modFire(n *Node) error {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]

	atmp, btmp, same := Promote(n, a.SrcGet(), b.SrcGet())

	if !same {
		n.LogError("incompatible types for modulo (%v%%%v)", reflect.TypeOf(a.Val), reflect.TypeOf(b.Val))
		x.DstPut(nil)
	} else {
		x.DstPut(modFire2(atmp, btmp))
	}
	return nil
}

// FuncMod is the module operator (x = a % b).
func FuncMod(a, b, x Edge) Node {

	node := MakeNode("mod", []*Edge{&a, &b}, []*Edge{&x}, nil, modFire)
	return node
}
