package imagelab

import (
	"testing"
	"time"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/vectaport/flowgraph"
)

var images = []string{"airplane.jpg", "fruits.jpg", "pic1.png", "pic3.png", "pic5.png", "stuff.jpg",
	"baboon.jpg", "lena.jpg", "pic2.png", "pic4.png", "pic6.png"}

func tbi(x flowgraph.Edge) flowgraph.Node {

	node := flowgraph.MakeNode("tbi", nil, []*flowgraph.Edge{&x}, nil, 
		func (n *flowgraph.Node) { 
			filename := "../../../lazywei/go-opencv/images/"+images[n.Cnt%int64(len(images))]
			n.Tracef("Loading %s\n", filename)
			x.Val = opencv.LoadImage(filename)
		})
	return node
}

func TestDisplay(t *testing.T) {

	test := true

	var quitChan chan flowgraph.Nada
	var wait time.Duration
	if !test {
		quitChan =make(chan flowgraph.Nada)
	} else {
		wait = 1
	}

	e,n := flowgraph.MakeGraph(1,2)
 
	n[0] = tbi(e[0])
	n[1] = FuncDisplay(e[0], quitChan)

	flowgraph.TraceLevel = flowgraph.V
	flowgraph.RunAll(n, time.Duration(wait*time.Second))

	if !test {
		<- quitChan
	}
}

