package imagelab

import (
	"github.com/lazywei/go-opencv/opencv"
	"github.com/vectaport/flowgraph"
)

type displayStruct struct {
	window *opencv.Window
	quitChan chan flowgraph.Nada
}

	
func displayFire (n *flowgraph.Node) {


	a := n.Srcs[0]

	window := a.Aux.(displayStruct).window
	image := a.Val.(*opencv.IplImage)
	defer image.Release()

	window.ShowImage(image)
	if a.Aux.(displayStruct).quitChan != nil {
		key := opencv.WaitKey(0)
		if key == 27 {
			var nada flowgraph.Nada
			a.Aux.(displayStruct).quitChan <- nada
		}
	} else {
		_ = opencv.WaitKey(1)
	}

}

// FuncDisplay displays an opencv image.
func FuncDisplay(a flowgraph.Edge, quitChan chan flowgraph.Nada) flowgraph.Node {
	node := flowgraph.MakeNode("display", []*flowgraph.Edge{&a}, nil, nil, displayFire)
	a.Aux = displayStruct{opencv.NewWindow("display"), quitChan}
	return node
}

