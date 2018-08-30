package imglab

import (
	"testing"
	"time"

	"github.com/vectaport/fgbase"
)

func TestCapture(t *testing.T) {

	test := true

	var quitChan chan struct{}
	var wait time.Duration
	if !test {
		quitChan = make(chan struct{})
	} else {
		wait = 1
	}

	e,n := fgbase.MakeGraph(1,2)
 
	n[0] = FuncCapture(e[0])
	n[1] = FuncDisplay(e[0], quitChan)

	fgbase.TraceLevel = fgbase.V
	fgbase.RunAll(n, time.Duration(wait*time.Second))

	if !test {
		<- quitChan
	}

}

