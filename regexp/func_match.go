package regexp

import (
	"github.com/vectaport/flowgraph"
)

func backslash(str string) func() (char byte, bslashed bool) {
	s := str
	return func() (char byte, bslashed bool) {
		if len(s)==0 {return 0x00,false}
		c := s[0]
		if c!='\\' || len(s)==1 {
			s = s[1:]
			return c,false
		} else {
			c = s[1]
			s = s[2:]
			return c,true
		}
	}
}

func matchFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	b := n.Srcs[1] 		 
	x := n.Dsts[0]

	av := a.SrcGet()
	bv := b.SrcGet()
	pattern := bv.(string)

        if av.(Search).State==Fail {
                x.DstPut(av)
                return
        }

	orig := av.(Search).Orig
	curr := av.(Search).Curr

	bssf := backslash(curr)
	bspf := backslash(pattern)
	
	matched := true
	for {

		// pattern is done
		pcurr,pbs := bspf()
		if pcurr == 0x00 {
			break
		} 

		// string is done
		scurr,_ := bssf()
		if scurr == 0x00 {
			matched = false
			break
		} 

		// match is over
		if scurr != pcurr && (pcurr != '.' || pbs) {
			matched = false
			break
		}
		
        }
	if matched {
		x.DstPut(Search{Curr:curr[len(pattern):], Orig:orig, State:Live})
		return
	}

	x.DstPut(Search{Curr:curr, Orig:orig})
	return
}

// FuncMatch advances a byte slice if it matches a string, otherwise returns the empty slice
func FuncMatch(a,b flowgraph.Edge, x flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("match", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, matchFire)
	return node
	
}
	
