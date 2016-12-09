package regexp

import (
	"github.com/vectaport/flowgraph"
)

type starStruct struct {
	prev map[string]string
}

func starFire (n *flowgraph.Node) {
	newmatch := n.Srcs[0]
	subsrc := n.Srcs[1]
	dnstreq := n.Srcs[2]

	oldmatch := n.Dsts[0]
	subdst := n.Dsts[1]
	upstreq := n.Dsts[2]

	st := n.Aux.(starStruct)

	if dnstreq.SrcRdy(n) {


		// match >0
		match := dnstreq.Val.(Search)
		if match.State==Fail {
			delete(st.prev, match.Orig)
			subdst.NoOut = true
		} else {
			match.Curr = st.prev[match.Orig]
			subdst.Val = match
		}
		
		newmatch.NoOut = true
		subsrc.NoOut = true
		oldmatch.NoOut = true
		upstreq.NoOut = true
		return
	}

	if subsrc.SrcRdy(n) {

		// match >0
		newmatch.Val = subsrc.Val
			
		newmatch.NoOut = true
		dnstreq.NoOut = true
		subdst.NoOut = true
		upstreq.NoOut = true
		return
	}

	if newmatch.SrcRdy(n) {

		// match zero
		match := newmatch.Val.(Search)
		st.prev[match.Orig]=match.Curr
		oldmatch.Val = match

		subsrc.NoOut = true
		dnstreq.NoOut = true
		subdst.NoOut = true
		upstreq.NoOut = true
		return
	}
	

}

func starRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) || !n.Dsts[2].DstRdy(n) { return false }
	return n.Srcs[0].SrcRdy(n) || n.Srcs[1].SrcRdy(n) || n.Srcs[2].SrcRdy(n)
}

// FuncStar repeats a match
//
// inputs:
// newmatch -- new match string
// subsrc   -- fedback result of last match, successful (remainder string) or not (nil)
// dnstreq  -- receive downstream request for new remainder string
//
// outputs:
// oldmatch -- continue match (remainder string)
// subdst   -- match done, successful (remainder string) or not (nil)
// upstreq  -- send upstream request for new remainder string
func FuncStar(newmatch, subsrc, dnstreq flowgraph.Edge, oldmatch, subdst, upstreq flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("star", []*flowgraph.Edge{&newmatch, &subsrc, &dnstreq}, []*flowgraph.Edge{&oldmatch, &subdst, &upstreq}, starRdy, starFire)
	node.Aux = starStruct{prev:make(map[string]string)}
	return node

}
