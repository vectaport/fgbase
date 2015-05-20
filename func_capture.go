package flowgraph

import (
	"github.com/lazywei/go-opencv/opencv"
	// "github.com/chai2010/opencv"
)

func captureFire (n *Node) {

	x := n.Dsts[0]
	cap := x.Aux.(*opencv.Capture)
	if cap.GrabFrame() {
		i1 := cap.RetrieveFrame(1)
		i2 := i1.Clone()
		if i2 == nil  {
			n.Tracef("image capture returned nil")
		}
		x.Val = i2
	}

}

func FuncCapture(x Edge) Node {
	node := MakeNode("capture", nil, []*Edge{&x}, nil, captureFire)

	x.Aux = opencv.NewCameraCapture(0)
	if x.Aux == nil {
		panic("cannot open capture device")
	}

	return node
}

