package flowgraph

import (
	"github.com/lazywei/go-opencv/opencv"
)

func displayFire (n *Node) {

	a := n.Srcs[0]

	window := a.Aux.(*opencv.Window)
	filename := a.Val.(string)
	image := opencv.LoadImage(filename)
	if image == nil {
		panic("LoadImage fail")
	}
	defer image.Release()

	window.ShowImage(image)
	n.Tracef("Displayed %s\n", filename)
}

func FuncDisplay(a Edge) Node {
	node := MakeNode("display", []*Edge{&a}, nil, nil, displayFire)

	a.Aux = opencv.NewWindow("display")

	return node
}

