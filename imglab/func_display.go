package imglab

import (
	"github.com/lazywei/go-opencv/opencv"
	"github.com/vectaport/fgbase"
)

type displayStruct struct {
	window *opencv.Window
	quitChan chan struct{}
}

	
func displayFire (n *fgbase.Node) {


	a := n.Srcs[0]

	window := n.Aux.(displayStruct).window
	image := a.SrcGet().(*opencv.IplImage)
	defer image.Release()

	window.ShowImage(image)
	if n.Aux.(displayStruct).quitChan != nil {
		key := opencv.WaitKey(0)
		if key == 27 {
			var nada struct{}
			n.Aux.(displayStruct).quitChan <- nada
		}
	} else {
		// _ = opencv.WaitKey(1)
	}

}

// FuncDisplay displays an opencv image.
func FuncDisplay(a fgbase.Edge, quitChan chan struct{}) fgbase.Node {
	node := fgbase.MakeNode("display", []*fgbase.Edge{&a}, nil, nil, displayFire)
	node.Aux = displayStruct{opencv.NewWindow("display"), quitChan}
	return node
}

