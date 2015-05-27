package flowgraph

import (		
	"net/http"
)      			

// FuncServer creates an http server and passes requests downstream.
func FuncServer(x Edge, addr string, quitChan chan Nada) Node {

	node := MakeNode("server", nil, []*Edge{&x}, nil, nil)

	http.HandleFunc("/count/", 
		func(w http.ResponseWriter, req *http.Request) {
			x := node.Dsts[0]
			x.Val = req
			node.FireThenWait()
		})
	node.RunFunc = func (n *Node) { 
		n.LogError("%v", http.ListenAndServe(addr, nil))
		var nada Nada
		quitChan <- nada
	}

	return node
}
	
