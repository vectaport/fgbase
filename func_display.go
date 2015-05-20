package flowgraph

import (
	"github.com/lazywei/go-opencv/opencv"
)

type displayStruct struct {
	window *opencv.Window
	quitChan chan Nada
}

	
func displayFire (n *Node) {


	a := n.Srcs[0]

	window := a.Aux.(displayStruct).window
	image := a.Val.(*opencv.IplImage)
	defer image.Release()

	window.ShowImage(image)
	if a.Aux.(displayStruct).quitChan != nil {
		key := opencv.WaitKey(0)
		if key == 27 {
			var nada Nada
			a.Aux.(displayStruct).quitChan <- nada
		}
	} else {
		_ = opencv.WaitKey(1)
	}

}

func FuncDisplay(a Edge, quitChan chan Nada) Node {
	node := MakeNode("display", []*Edge{&a}, nil, nil, displayFire)
	a.Aux = displayStruct{opencv.NewWindow("display"), quitChan}
	return node
}

