package flowgraph

import (		
	"encoding/csv"
	"io"
	"os"
)      			

func check(e error) {
	if e != nil {
		StderrLog.Printf("%v\n", e)
		os.Exit(1)
	}
}
		
func csviFire (n *Node) {	 
	x := n.Dsts
	r := n.Aux.(*csv.Reader)
	var err error

	// read data string
	record, err := r.Read()
	check(err)
	l := len(x)
	if l>len(record) { l = len(record) }
	for i:=0; i<l; i++ {
		if record[i]!="*" {
        		v := ParseDatum(record[i])
			x[i].Val = v	
		} else {
			x[i].NoOut = true
		}
	}
}

// FuncCSVI reads a vector of input data values from a Reader.
// 
func FuncCSVI(x []Edge, r io.Reader) Node {

	var xp []*Edge
	for i := range x {
		xp = append(xp, &x[i])
	}

	node := MakeNode("csvi", nil, xp, nil, csviFire)
	r2 := csv.NewReader(r)
	node.Aux = r2

	// skip headers
	_, err := r2.Read()
	check(err)

	return node
	
}
	
