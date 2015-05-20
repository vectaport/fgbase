package flowgraph

import (
	"testing"
	"time"
)

func TestCapture(t *testing.T) {

	test := true

	var quitChan chan Nada
	var wait time.Duration
	if !test {
		quitChan = make(chan Nada)
	} else {
		wait = 1
	}

	e,n := MakeGraph(1,2)
 
	n[0] = FuncCapture(e[0])
	n[1] = FuncDisplay(e[0], quitChan)

	TraceLevel = V
	RunAll(n, time.Duration(wait*time.Second))

	if !test {
		<- quitChan
	}

}

