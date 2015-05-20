package flowgraph

import (
	"fmt"

	"github.com/lazywei/go-opencv/opencv"
)

func smoothFire (n *Node) {

	a := n.Srcs[0]
	x := n.Dsts[0]	

	if a.Val != nil {
		img0 := a.Val.(*opencv.IplImage)
		defer img0.Release()
		
		img1 := img0.Clone()
		opencv.Smooth(img0, img1, opencv.CV_BLUR, 3, 3, 0, 0)

		if TraceLevel > Q {
			for i:=0; i<3; i++ {
				var s string = "BEFORE: "
				for j:=0; j<3; j++ {
					s += fmt.Sprintf("%v ", img0.Get2D(100+j, 100+i))
				}
				n.Tracef(s+"\n")
			}
			
			for i:=0; i<3; i++ {
				var s string = "AFTER: "
				for j:=0; j<3; j++ {
					s += fmt.Sprintf("%v ", img1.Get2D(100+j, 100+i))
				}
				n.Tracef(s+"\n")
			}
		}
		
		x.Val = img1
	} else {
		x.Val = nil
	}

}

// FuncSmooth smoothes an opencv image.
func FuncSmooth(a, x Edge) Node {
	node := MakeNode("smooth", []*Edge{&a}, []*Edge{&x}, nil, smoothFire)
	return node
}

