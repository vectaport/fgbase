package regexp

import (		
       "github.com/vectaport/flowgraph"
)      			

func barFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	b := n.Srcs[1] 		 
	x := n.Dsts[0]
	sink := n.Aux.(bool)

        if a.Val==nil && b.Val==nil {
		if sink {
			x.NoOut = true
			return
		}
                x.Val = nil
                return
        }

        if a.Val==nil {
	        x.Val = b.Val
		return
        }

        if b.Val==nil {
	        x.Val = a.Val
		return
        }

        as := a.Val.(Regexp).Curr
        bs := b.Val.(Regexp).Curr
	if len(as)>len(bs) {
  		x.Val = a.Val
        }
        x.Val = b.Val
}

// FuncBar waits for both inputs and returns the one that matches the shortest string.
// Returns nil if no match at all.
func FuncBar(a,b flowgraph.Edge, x flowgraph.Edge, sink bool) flowgraph.Node {
	
	node := flowgraph.MakeNode("bar", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, barFire)
	node.Aux = sink
	return node
	
}
	
