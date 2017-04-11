package regexp

import (
	"github.com/vectaport/flowgraph"
)

func preprocess(str string) func() (char byte, bslashed bool) {
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

	stringFunc := preprocess(curr)
	patternFunc := preprocess(pattern)
	
	matched := true
	pcnt := 0
	for {

		// if pattern is done
		pcurr,pbs := patternFunc()
		if pcurr == 0x00 {
			break
		} 
		pcnt++

		// if string is done
		scurr,_ := stringFunc()
		if scurr == 0x00 {
			matched = false
			break
		}

		// if pattern is a bracket list
		if pcurr=='[' {
			for {
				ch,_ := patternFunc()
				if ch == ']' {
					break
				}
				if ch == scurr {
					pcurr = ch
				}
			}
		}

		// match is over
		if scurr != pcurr && (pcurr != '.' || pbs) {
			matched = false
			break
		}
		
        }
	if matched {
		x.DstPut(Search{Curr:curr[pcnt:], Orig:orig})
		return
	}

	x.DstPut(Search{Curr:curr, Orig:orig, State:Fail})
	return
}

// FuncMatch advances a byte slice if it matches a string, otherwise returns the empty slice
func FuncMatch(a,b flowgraph.Edge, x flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("match", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, matchFire)
	return node
	
}
	
