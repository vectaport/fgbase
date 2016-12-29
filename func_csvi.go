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

// FuncCSVI reads a vector of input data values from a Reader and outputs
// them downstream.  enums is an optional map from field.enum to an empty interface.
func FuncCSVI(x []Edge, r io.Reader, enums map[string]interface{}) Node {

	var fireFunc = func (n *Node) {	 
		x := n.Dsts
		
		// process data record
		record := n.Aux.(csvState).record
		header := n.Aux.(csvState).header
		l := len(x)
		if l>len(record) { l = len(record) }
		for i:=0; i<l; i++ {
			j := header[i]
			// n.Tracef("i=%d, j=%d\n", i, j)
			// n.Tracef("record=%v, header=%v\n", record, header)
			if j>=0 {
				if record[j]=="*" {
					continue
				}
				var v interface{}
				var ok bool
				if enums!= nil {
					v,ok = enums[record[j]]
				}
				if !ok {
						v = ParseDatum(record[j])
				}
				x[i].DstPut(v)
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
	for i := range x {
		ix := find(x[i].Name, headers)
		if ix>= 0 {
			h = append(h, ix)
		}
	}
	node.Aux = csvState{csvreader:r2, header:h}

	return node
	
}
	
