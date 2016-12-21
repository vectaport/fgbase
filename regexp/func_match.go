package regexp

import (
	"github.com/vectaport/flowgraph"
)

func backslash(str string) func() (char byte, bquoted bool) {
	s := str
	return func() (char byte, bquoted bool) {
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
	match := b.Val.(string)

        if a.Val.(Search).State==Fail {
                x.Val = a.Val
                return
        }

	orig := a.Val.(Search).Orig
	curr := a.Val.(Search).Curr

	bssf := backslash(curr)
	bsmf := backslash(match)
	
	matched := true
	for {
		scurr,_ := bssf()
		mcurr,mbs := bsmf()
		
		if mcurr == 0x00 { break } // match is done
		
		if scurr == 0x00 { matched = false; break } // string is done
		
		if scurr != mcurr && (mcurr != '.' || mbs) { matched = false; break } // match is over
        }
	if matched {
		x.Val = Search{Curr:curr[len(match):], Orig:orig, State:Live}
		return
	}

	x.Val = Search{Curr:curr, Orig:orig}
	return
}

// FuncMatch advances a byte slice if it matches a string, otherwise returns the empty slice
func FuncMatch(a,b flowgraph.Edge, x flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("match", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, matchFire)
	return node
	
}
	
