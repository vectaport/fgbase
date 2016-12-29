package flowgraph

import (
	"encoding/csv"
	"io"
	"os"
	"reflect"
)      			


func csvoRdy (n *Node) bool {
	if n.Aux == nil { return false }
	
	a := n.Srcs
	r := n.Aux.(csvState).csvreader
	h := n.Aux.(csvState).header

	if n.Aux== nil { return false }

	if n.Aux.(csvState).record==nil {
		record,err := r.Read()
		if err == io.EOF {
			os.Exit(0)
			return false
		} else {
			check(err)
			n.Aux = csvState{r, h, record}
		}
	}

	record := n.Aux.(csvState).record
	header := n.Aux.(csvState).header

	for i := range a {
		j := header[i]
		if j>= 0 {
			if !a[i].SrcRdy(n) {
				if record[j]!="*" {
					return false
				}
			}
		} else {
			n.LogError("Named output missing from .csv file:  %s\n", a[i].Name)
			os.Exit(1)
		}
	}
	return true
}

// FuncCSVO reads a vector of expected data values from a Reader and tests the against
// input from upstream.  enums is an optional map from field.enum to an empty interface.
func FuncCSVO(a []Edge, r io.Reader, enums map[string]interface{}) Node {

	var fireFunc = func (n *Node) {	 
		a := n.Srcs
		
		r := n.Aux.(csvState).csvreader
		header := n.Aux.(csvState).header
		record := n.Aux.(csvState).record
		
		l := len(a)
		if l>len(record) { l = len(record) }
		for i:=0; i<l; i++ {
			j := header[i]

			if record[j]!="*" {
				var v interface{}
				var ok bool
				if enums!=nil { 
					v,ok = enums[record[j]]
				}
				if !ok {
					v = ParseDatum(record[j])
				}
				av := a[i].SrcGet()
				if !EqualsTest(n, v, av) {

					// check if space-separated struct to compare
					if record[j][0]=='{' && record[j][len(record[j])-1]=='}' && IsStruct(av) {
						l := len(record[j])
						s := record[j][1:l-1]
						m := 0
						for {
							if s == "" { break }
							var p string
							k := 0
							for {
								if k==len(s) || s[k]==' ' { break }
								p += string(s[k])
								k++
							}
							if k!=len(s) {
								s = s[k+1:]
							} else {
								s = ""
							}
							if enums != nil {
								v,ok = enums[p]
							}
							if !ok {
								v = ParseDatum(p)
							}
							
							av := reflect.ValueOf(av)
							ft := av.Field(m).Interface()
							if !EqualsTest(n, v, ft) {
								n.LogError("%s:  expected %T(%v|0x%x) from field %d of %T(%v)", 
									a[i].Name, v, v, v, i, av, av)	
							}
							m++
						}
					} else {
						n.LogError("%s:  expected %T(%v|0x%x), actual %T(%v|0x%x)", 
							a[i].Name, v, v, v, av, av, av)
					}
				}
			}
		}
		
		n.Aux = csvState{csvreader:r, header:header}
		
	}
	
	var ap []*Edge
	for i := range a {
		ap = append(ap, &a[i])
	}

	node := MakeNode("csvo", ap, nil, csvoRdy, fireFunc)
	r2 := csv.NewReader(r)

	// save headers
	headers, err := r2.Read()
	check(err)
	var h []int
	for i := range a {
		h = append(h, find(a[i].Name, headers))
	}
	node.Aux = csvState{csvreader:r2, header:h}

	return node
	
}
	
