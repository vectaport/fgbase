package weblab

import (
	"fmt"
	"net/http"

	"github.com/vectaport/fgbase"
)

type handler struct {
	subhandle func(http.ResponseWriter, *http.Request)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.subhandle(w, r)
}

// FuncHTTP creates an http server and passes requests downstream.
func FuncHTTP(x fgbase.Edge, addr string, quitChan chan struct{}) fgbase.Node {

	node := fgbase.MakeNode("http", nil, []*fgbase.Edge{&x}, nil, nil)

	var h = &handler{
		func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, ".")
			x := node.Dsts[0]
			x.DstPut(req.URL)
			node.TraceValRdy()
			if node.RdyAll() {
				x.SendData(&node)
			}
		},
	}

	node.RunFunc = func(n *fgbase.Node) {
		n.LogError("%v", http.ListenAndServe(addr, h))
		var nada struct{}
		quitChan <- nada
	}

	return node
}
