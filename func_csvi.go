package flowgraph

import (		
	"encoding/csv"
	"io"
)

type readerrecord struct {
	csvreader *csv.Reader
	record []string
}

func csviRdy (n *Node) bool {
	if n.Aux == nil { return false }
	
	if n.DefaultRdyFunc() {
		r := n.Aux.(readerrecord).csvreader
		record,err := r.Read()
		if err == io.EOF {
			n.Aux = nil
			return false
		} else {
			check(err)
			n.Aux = readerrecord{r, record}
			return true
		}
	}
	return false
}

func csviFire (n *Node) {	 
	x := n.Dsts

	// read data string
	record := n.Aux.(readerrecord).record
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

	node := MakeNode("csvi", nil, xp, csviRdy, csviFire)
	r2 := csv.NewReader(r)
	node.Aux = readerrecord{csvreader:r2}

	// skip headers
	_, err := r2.Read()
	check(err)

	return node
	
}
	
