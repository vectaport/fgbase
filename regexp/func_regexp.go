package regexp

import (
	"github.com/vectaport/flowgraph"
)

type regexpStruct struct {
	live bool
	prev string
}

func regexpFire (n *flowgraph.Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]

	live := n.Aux.(regexpStruct).live
	prev := n.Aux.(regexpStruct).prev

	// first attempted match
	if !live {
		b.NoOut = true
		if a.Val == nil {
			x.NoOut = true
			y.Val = nil
			return
		}
		x.Val = a.Val
		y.NoOut = true
		curr := a.Val.(Regexp).Curr
		if curr!="" {
			n.Aux = regexpStruct{live:true, prev:curr}
			return
		}
		n.Aux = regexpStruct{live:true}
		return
	}
	
	a.NoOut = true
	
	// if regexp failed advance string if any left
	if b.Val==nil {
		if len(prev)<=1 {
			x.NoOut = true
			y.Val = nil
			n.Aux = regexpStruct{live:false}
			return
		}
		curr := prev[1:]
		x.Val = Regexp{Curr:curr, Orig:"LOSTORIG2"}
		y.NoOut = true
		n.Aux = regexpStruct{live:true, prev:curr}
		return
	}
	
	x.NoOut = true
	y.Val = b.Val
	n.Aux = regexpStruct{live:false}
	return
}

func regexpRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) { return false }
	live := n.Aux.(regexpStruct).live
	if live { return n.Srcs[1].SrcRdy(n) }
	return n.Srcs[0].SrcRdy(n)
}

// FuncRegexp repeats a match
// inputs:
// a -- new string
// b -- fedback result of last regexp, successful (remainder string) or not (nil)
// outputs:
// x -- continue regexp (remainder string)
// y -- regepx done, successful (remainder string) or not (nil)
func FuncRegexp(a, b, x, y flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("regexp", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x, &y}, regexpRdy, regexpFire)
	node.Aux = regexpStruct{}
	return node

}
