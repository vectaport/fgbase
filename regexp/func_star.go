package regexp

import (
	"github.com/vectaport/flowgraph"
)

type starStruct struct {
	live bool
	prev string
}

func starFire (n *flowgraph.Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]

	live := n.Aux.(starStruct).live
	prev := n.Aux.(starStruct).prev

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
		s,ok := a.Val.(string)
		if ok {
			n.Aux = starStruct{live:true, prev:s}
			return
		}
		n.Aux = starStruct{live:true}
		return
	}
	
	a.NoOut = true
	
	// if match failed
	if b.Val==nil {
		x.NoOut = true
		y.Val = prev
		n.Aux = starStruct{live:false, prev:prev}
		return
	}
	
	s := b.Val.(string)
	
	// match is complete
	if len(s)==0 {
		x.NoOut = true
		y.Val = s
		n.Aux = starStruct{live:false}
		return
	}
	
	// if match goes on
	x.Val = s
	y.NoOut = true
	n.Aux = starStruct{live:true, prev:b.Val.(string)}
	return
}

func starRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) { return false }
	live := n.Aux.(starStruct).live
	if live { return n.Srcs[1].SrcRdy(n) }
	return n.Srcs[0].SrcRdy(n)
}

// FuncStar repeats a match
// inputs:
// a -- new match (string)
// b -- fedback result of last match, successful (remainder string) or not (nil)
// outputs:
// x -- continue match (remainder string)
// y -- match done, successful (remainder string) or not (nil)
func FuncStar(a, b, x, y flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("star", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x, &y}, starRdy, starFire)
	node.Aux = starStruct{}
	return node

}
