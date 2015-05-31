package flowgraph

import (		
	"fmt"
	"net/http"
)      			

type handler struct {subhandle func(http.ResponseWriter, *http.Request)}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.subhandle(w, r)
}

// FuncHttp creates an http server and passes requests downstream.
func FuncHttp(x Edge, addr string, quitChan chan Nada) Node {

	node := MakeNode("http", nil, []*Edge{&x}, nil, nil)

	var h = &handler{
		func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, ".")
			x := node.Dsts[0]
			x.Val = req.URL
			if TraceLevel >= VVV {node.traceValRdy(false)}
			if node.RdyAll() {
				x.SendData(&node)
			}
		},
	}

	node.RunFunc = func (n *Node) { 
		n.LogError("%v", http.ListenAndServe(addr, h))
		var nada Nada
		quitChan <- nada
	}

	return node
}
	
