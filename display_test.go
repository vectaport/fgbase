package flowgraph

import (
	"testing"
	"time"

	"github.com/lazywei/go-opencv/opencv"
)

var images = []string{"airplane.jpg", "fruits.jpg", "pic1.png", "pic3.png", "pic5.png", "stuff.jpg",
	"baboon.jpg", "lena.jpg", "pic2.png", "pic4.png", "pic6.png"}

func tbi(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) { 
			filename := "../../lazywei/go-opencv/images/"+images[n.Cnt%int64(len(images))]
			n.Tracef("Loading %s\n", filename)
			x.Val = opencv.LoadImage(filename)
		})
	return node
}

func TestDisplay(t *testing.T) {

	test := true

	var quitChan chan Nada
	var wait time.Duration
	if !test {
		quitChan =make(chan Nada)
	} else {
		wait = 1
	}

	e,n := MakeGraph(1,2)
 
	n[0] = tbi(e[0])
	n[1] = FuncDisplay(e[0], quitChan)

	TraceLevel = V
	RunAll(n, time.Duration(wait*time.Second))

	if !test {
		<- quitChan
	}
}

