package flowgraph

import (		
	"encoding/csv"
	"io"
	"os"
)

type csvState struct {
	csvreader *csv.Reader
	header []int // index of input/output based on header
	record []string
}

func find(s string, v []string) int {
	for i := range v {
		if v[i]==s {
			return i
		}
	}
	return -1
}

func csviRdy (n *Node) bool {
	if n.Aux == nil { return false }
	
	if n.DefaultRdyFunc() {
		r := n.Aux.(csvState).csvreader
		header := n.Aux.(csvState).header
		record,err := r.Read()
		if err == io.EOF {
			n.Aux = nil
			return false
		} else {
			check(err)
			n.Aux = csvState{r, header, record}
			return true
		}
	}
	return false
}

// FuncCSVI reads a vector of input data values from a Reader.
func FuncCSVI(x []Edge, r io.Reader, enums map[string]int) Node {

	var fireFunc = func (n *Node) {	 
		x := n.Dsts
		
		// process data record
		record := n.Aux.(csvState).record
		header := n.Aux.(csvState).header
		l := len(x)
		if l>len(record) { l = len(record) }
		for i:=0; i<l; i++ {
			j := header[i]
			if j>=0 {
				if record[j]=="*" {
					x[i].NoOut = true
					continue
				}
				var v Datum
				var ok bool
				v,ok = enums[record[j]]
				if !ok {
					v = ParseDatum(record[j])
				}
				x[i].Val = v	
			} else {
				n.LogError("Named input missing from .csv file:  %s\n", x[i].Name)
				os.Exit(1)
			}
		}
	}

	var xp []*Edge
	for i := range x {
		xp = append(xp, &x[i])
	}

	node := MakeNode("csvi", nil, xp, csviRdy, fireFunc)
	r2 := csv.NewReader(r)

	// save headers
	headers, err := r2.Read()
	check(err)
	var h []int
	for i := range headers {
		h = append(h, find(x[i].Name, headers))
	}
	node.Aux = csvState{csvreader:r2, header:h}

	return node
	
}
	
