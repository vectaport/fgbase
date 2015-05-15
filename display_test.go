package flowgraph

import (
	"testing"
	"time"
)

var images = []string{"airplane.jpg", "fruits.jpg", "pic1.png", "pic3.png", "pic5.png", "stuff.jpg",
	"baboon.jpg", "lena.jpg", "pic2.png", "pic4.png", "pic6.png"}

func tbiDisplay(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) { 
			x.Val = "../../lazywei/go-opencv/images/"+images[n.Cnt%int64(len(images))]
		})
	return node
}

func TestDisplay(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(1,2)
 
	n[0] = tbiDisplay(e[0])
	n[1] = FuncDisplay(e[0])

	RunAll(n, time.Second)

}

