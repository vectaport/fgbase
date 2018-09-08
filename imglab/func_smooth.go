package imglab

import (
	"fmt"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/vectaport/fgbase"
)

func smoothFire(n *fgbase.Node) error {

	a := n.Srcs[0]
	x := n.Dsts[0]

	if a.Val != nil {
		img0 := a.Val.(*opencv.IplImage)
		defer img0.Release()

		img1 := img0.Clone()
		opencv.Smooth(img0, img1, opencv.CV_BLUR, 3, 3, 0, 0)

		if fgbase.TraceLevel > fgbase.Q {
			for i := 0; i < 3; i++ {
				var s = "BEFORE: "
				for j := 0; j < 3; j++ {
					s += fmt.Sprintf("%v ", img0.Get2D(100+j, 100+i))
				}
				n.Tracef(s + "\n")
			}

			for i := 0; i < 3; i++ {
				var s = "AFTER: "
				for j := 0; j < 3; j++ {
					s += fmt.Sprintf("%v ", img1.Get2D(100+j, 100+i))
				}
				n.Tracef(s + "\n")
			}
		}

		x.Val = img1
	} else {
		x.Val = nil
	}
	return nil

}

// FuncSmooth smoothes an opencv image.
func FuncSmooth(a, x fgbase.Edge) fgbase.Node {
	node := fgbase.MakeNode("smooth", []*fgbase.Edge{&a}, []*fgbase.Edge{&x}, nil, smoothFire)
	return node
}
