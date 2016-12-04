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
		curr := a.Val.(Regexp).Curr
		if curr!="" {
			n.Aux = starStruct{live:true, prev:curr}
			return
		}
		n.Aux = starStruct{live:true}
		return
	}
	
	a.NoOut = true
	
	// if match failed
	if b.Val==nil {
		x.NoOut = true
		y.Val = Regexp{Curr:prev, Orig:"LOSTORIG"}
		n.Aux = starStruct{live:false, prev:prev}
		return
	}
	
	curr := b.Val.(Regexp).Curr
	orig := b.Val.(Regexp).Orig
	
	// match is complete
	if len(curr)==0 {
		x.NoOut = true
		y.Val = Regexp{Curr:curr, Orig:orig}
		n.Aux = starStruct{live:false}
		return
	}
	
	// if match goes on
	x.Val = Regexp{Curr:curr, Orig:orig}
	y.NoOut = true
	n.Aux = starStruct{live:true, prev:b.Val.(Regexp).Curr}
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
