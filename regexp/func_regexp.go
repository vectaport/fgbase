package regexp

import (
	"github.com/vectaport/flowgraph"
)

type regexpStruct struct {
	prev map[string]string
}

func regexpFire (n *flowgraph.Node) {
	newmatch := n.Srcs[0]
	subsrc := n.Srcs[1]
	dnstreq := n.Srcs[2]

	// oldmatch := n.Dsts[0]
	subdst := n.Dsts[1]
	// upstreq := n.Dsts[2]

	st := n.Aux.(regexpStruct)

	if dnstreq.SrcRdy(n) {


		// match >0
		match := dnstreq.SrcGet().(Search)
		if match.State==Fail {
			delete(st.prev, match.Orig)
		} else {
			match.Curr = st.prev[match.Orig]
			subdst.DstPut(match)
		}
		
		return
	}

	if subsrc.SrcRdy(n) {

		newmatch.DstPut(subsrc.SrcGet())
		return
		
	}

	if newmatch.SrcRdy(n) {

		match := newmatch.SrcGet().(Search)
		st.prev[match.Orig]=match.Curr
		subdst.DstPut(match)
		return
	}
	

}

func regexpRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) || !n.Dsts[2].DstRdy(n) { return false }
	return n.Srcs[0].SrcRdy(n) || n.Srcs[1].SrcRdy(n) || n.Srcs[2].SrcRdy(n)
}

// FuncRegexp does a match once
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
func FuncRegexp(newmatch, subsrc, dnstreq flowgraph.Edge, oldmatch, subdst, upstreq flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("regexp", []*flowgraph.Edge{&newmatch, &subsrc, &dnstreq}, []*flowgraph.Edge{&oldmatch, &subdst, &upstreq}, regexpRdy, regexpFire)
	node.Aux = regexpStruct{prev:make(map[string]string)}
	return node

}
