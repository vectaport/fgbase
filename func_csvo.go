package flowgraph

import (
	"encoding/csv"
	"io"
	"os"
)      			


func csvoRdy (n *Node) bool {
	if n.Aux == nil { return false }
	
	a := n.Srcs
	r := n.Aux.(readerrecord).csvreader

	if n.Aux== nil { return false }

	if n.Aux.(readerrecord).record==nil {
		record,err := r.Read()
		if err == io.EOF {
			os.Exit(0)
			return false
		} else {
			check(err)
			n.Aux = readerrecord{r, record}
		}
	}

	record := n.Aux.(readerrecord).record

	for i := range a {
		if !a[i].SrcRdy(n) {
			if record[i]!="*" {
				return false
			} else {
				a[i].NoOut = true
			}
		}
	}
	return true
}

func csvoFire (n *Node) {	 
	a := n.Srcs

	record := n.Aux.(readerrecord).record
	r := n.Aux.(readerrecord).csvreader

	l := len(a)
	if l>len(record) { l = len(record) }
	for i:=0; i<l; i++ {
		if record[i]!="*" {
			v := ParseDatum(record[i])
			if !EqualsTest(n, v, a[i].Val) {
				n.LogError("expected=%v, actual=%v", v, a[i].Val)	
			}
		}
	}

	n.Aux = readerrecord{csvreader:r}
	
}

// FuncCSVO reads a vector of input data values from a Reader.
func FuncCSVO(a []Edge, r io.Reader) Node {

	var ap []*Edge
	for i := range a {
		ap = append(ap, &a[i])
	}

	node := MakeNode("csvo", ap, nil, csvoRdy, csvoFire)
	r2 := readerrecord{csvreader:csv.NewReader(r)}
	node.Aux = r2

	// skip headers
	_, err := r2.csvreader.Read()
	check(err)

	return node
	
}
	
