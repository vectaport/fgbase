package flowgraph

import (		
	"encoding/csv"
	"io"
)      			

func csvoFire (n *Node) {	 
	a := n.Srcs
	r := n.Aux.(*csv.Reader)
	var err error

	// read data string
	record, err := r.Read()
	check(err)
	l := len(a)
	if l>len(record) { l = len(record) }
	for i:=0; i<l; i++ {
		if record[i]!="*" {
        		v := ParseDatum(record[i])
			if !EqualsTest(n, v, a[i].Val) {
				n.LogError("expected=%v, actual=%v", v, a[i].Val)	
			}
		} else {
			a[i].NoOut = true
		}
	}
}

// FuncCSVO reads a vector of input data values from a Reader.
// 
func FuncCSVO(a []Edge, r io.Reader) Node {

	var ap []*Edge
	for i := range a {
		ap = append(ap, &a[i])
	}

	node := MakeNode("csvo", ap, nil, nil, csvoFire)
	r2 := csv.NewReader(r)
	node.Aux = r2

	// skip headers
	_, err := r2.Read()
	check(err)

	return node
	
}
	
