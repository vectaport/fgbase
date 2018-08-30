package imglab

import (
	"github.com/lazywei/go-opencv/opencv"
	"github.com/vectaport/fgbase"
)

func captureFire (n *fgbase.Node) {

	x := n.Dsts[0]
	cap := n.Aux.(*opencv.Capture)
	if cap.GrabFrame() {
		i1 := cap.RetrieveFrame(1)
		i2 := i1.Clone()
		if i2 == nil  {
			n.Tracef("image capture returned nil")
		}
		x.DstPut(i2)
	}

}

// FuncCapture captures an opencv image.
func FuncCapture(x fgbase.Edge) fgbase.Node {
	node := fgbase.MakeNode("capture", nil, []*fgbase.Edge{&x}, nil, captureFire)

	node.Aux = opencv.NewCameraCapture(0)
	if node.Aux == nil {
		panic("cannot open capture device")
	}

	return node
}

