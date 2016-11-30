package regexp

import (
	"github.com/vectaport/flowgraph"
)

func bq(str string) func() (char byte, bquoted bool) {
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
	x := n.Dsts[0] 		 
	match := n.Aux.(string)


        if a.Val==nil {
                x.Val = nil
                return
        }

	s := a.Val.(string)

	bqsf := bq(s)
	bqmf := bq(match)
	
	matched := true
	for {
		scurr,_ := bqsf()
		mcurr,mbq := bqmf()
		
		if mcurr == 0x00 { break } // match is done
		
		if scurr == 0x00 { matched = false; break } // string is done
		
		if scurr != mcurr && (mcurr != '.' || mbq) { matched = false; break } // match is over
        }
	if matched {
		x.Val = s[len(match):]
		return
	}

	x.Val = nil
	return
}

// FuncMatch advances a byte slice if it matches a string, otherwise returns the empty slice
func FuncMatch(a flowgraph.Edge, x flowgraph.Edge, match string) flowgraph.Node {
	
	node := flowgraph.MakeNode("match", []*flowgraph.Edge{&a}, []*flowgraph.Edge{&x}, nil, matchFire)
	node.Aux = match
	return node
	
}
	
