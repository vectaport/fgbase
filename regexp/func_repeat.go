package regexp

import (
	"github.com/vectaport/flowgraph"
)

type repeatStruct struct {
	entries map[string]*repeatEntry
	rdy[3] bool
}

type repeatEntry struct {
	prev string
	min, max, cnt int
}

type repeatMap map[string]repeatEntry

func repeatFire (n *flowgraph.Node) {

	newmatch := n.Srcs[0]
	subsrc := n.Srcs[1]
	dnstreq := n.Srcs[2]

	oldmatch := n.Dsts[0]
	subdst := n.Dsts[1]
	// upstreq := n.Dsts[2]

	st := n.Aux.(repeatStruct)
	rmap := st.entries
	rdyv := st.rdy

	if rdyv[2] /* dnstreq */ {

		// match >0
		match := dnstreq.SrcGet().(Search)
		if match.State==Done {
			delete(rmap, match.Orig)
		} else {
			p := rmap[match.Orig].prev
			if len(p)==0 {
				n.Tracef("panic:  unable to send new string downstream\n");
				panic("unable to send new string downstream")
			}
			match.Curr = p[1:]
			subdst.DstPut(match)
		}
		
		return
	}

	if rdyv[1] /* subsrc */ {

		match := subsrc.SrcGet().(Search)
		rs := rmap[match.Orig]
		rs.prev = match.Curr
		rmap[match.Orig] = rs
		oldmatch.DstPut(match)
		return
		
	}

	if rdyv[0] /* newmatch */ {

		match := newmatch.SrcGet().(Search)
		rs := rmap[match.Orig]
		if rs==nil {
			rs = &repeatEntry{}
			rmap[match.Orig] = rs
		}
		rs.prev = match.Curr
		rmap[match.Orig] = rs

		// if no matches are required, pass it on
		if rs.min==0 {
			oldmatch.DstPut(match)
			return
		}

		// otherwise attempt a match
		subdst.DstPut(match)
		return


	}
	

}

func repeatRdy (n *flowgraph.Node) bool {
	if !n.Dsts[0].DstRdy(n) || !n.Dsts[1].DstRdy(n) || !n.Dsts[2].DstRdy(n) { return false }
	st := n.Aux.(repeatStruct)
	st.rdy[0] = n.Srcs[0].SrcRdy(n)
	st.rdy[1] =  n.Srcs[1].SrcRdy(n)
	st.rdy[2] = n.Srcs[2].SrcRdy(n)
	n.Aux = st
	return st.rdy[0] || st.rdy[1] || st.rdy[2]
}

// FuncRepeat repeats a match zero or more times
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
func FuncRepeat(newmatch, subsrc, dnstreq flowgraph.Edge, oldmatch, subdst, upstreq flowgraph.Edge, min, max int) flowgraph.Node {

	node := flowgraph.MakeNode("repeat", []*flowgraph.Edge{&newmatch, &subsrc, &dnstreq}, []*flowgraph.Edge{&oldmatch, &subdst, &upstreq}, repeatRdy, repeatFire)
	node.Aux = repeatStruct{entries:make(map[string]*repeatEntry)}
	return node

}
