package regexp

import (
	"github.com/vectaport/flowgraph"
)

type repeatStruct struct {
	entries map[string]*repeatEntry
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

	if dnstreq.Flow /* dnstreq */ {

		newmatch.Flow = false
		subsrc.Flow = false

		// match >0
		match := dnstreq.SrcGet().(Search)
		if match.State==Done {
			delete(rmap, match.Orig)
		} else {
			if rmap[match.Orig]==nil {
				n.Tracef("panic:  nil return from rmap for \"%+v\"  (%v)\n", match, rmap)
				panic("nil return from rmap")
			}
			
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

	if subsrc.Flow /* subsrc */ {

		newmatch.Flow = false
		dnstreq.Flow = false
		
		match := subsrc.SrcGet().(Search)
		rs := rmap[match.Orig]
		rs.prev = match.Curr
		rmap[match.Orig] = rs
		oldmatch.DstPut(match)
		return
		
	}

	if newmatch.Flow /* newmatch */ {

		subsrc.Flow = false
		dnstreq.Flow = false
		
		match := newmatch.SrcGet().(Search)
		rs := rmap[match.Orig]
		if rs==nil {
			rs = &repeatEntry{}
			rmap[match.Orig] = rs
		}
		rs.prev = match.Curr
		rmap[match.Orig] = rs
		n.Tracef("rmap after adding \"%s\":  %v\n", match.Orig, rmap)

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
	rdy := false
	for i := range n.Srcs {
		n.Srcs[i].Flow = n.Srcs[i].SrcRdy(n)
		rdy = rdy || n.Srcs[i].Flow
	}
	return rdy
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
