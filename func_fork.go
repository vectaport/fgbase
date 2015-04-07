package flowgraph

import (
	"reflect"
)

func forkFire (n *Node) { 
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = a.Val; 
	if IsSlice(a.Val) {
		at := reflect.TypeOf(a.Val)
		av := reflect.ValueOf(a.Val)
		y.Val = reflect.MakeSlice(at, av.Len(), av.Cap()).Interface()
		reflect.Copy(reflect.ValueOf(y.Val), reflect.ValueOf(a.Val))
	} else {
		y.Val = a.Val
	}
}

// FuncFork sends a value two ways (x = a; y = a)
func FuncFork(a, x, y Edge) {

	node := MakeNode("fork", []*Edge{&a}, []*Edge{&x, &y}, nil, forkFire)
	node.Run()
	
}
	
