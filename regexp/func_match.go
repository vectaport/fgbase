package regexp

import (		
       "github.com/vectaport/flowgraph"
)      			

func matchFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	x := n.Dsts[0] 		 
	match := n.Aux.(string)


        if a.Val==nil {
                x.Val = nil
                return
        }

	s := a.Val.(string)
	if len(s)<len(match) {
		x.Val = nil
		return
	}
	
	matched := true
	for i := range match {
		if s[i]!=match[i] && match[i] != '.' { matched = false; break }
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
	
