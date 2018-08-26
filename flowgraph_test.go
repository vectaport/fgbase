package flowgraph_test

import (
	"github.com/vectaport/flowgraph"
	"os"
	"testing"
)

/*=====================================================================*/

func TestMain(m *testing.M) {
	flowgraph.ConfigByFlag(nil)
	os.Exit(m.Run())
}

/*=====================================================================*/

func TestNewEqual(t *testing.T) {
	// Different allocations should not be equal.
	if flowgraph.New("abc") == flowgraph.New("abc") {
		t.Errorf(`New("abc") == New("abc")`)
	}
	if flowgraph.New("abc") == flowgraph.New("xyz") {
		t.Errorf(`New("abc") == New("xyz")`)
	}

	// Same allocation should be equal to itself (not crash).
	g := flowgraph.New("jkl")
	if g != g {
		t.Errorf(`graph != graph`)
	}
}

/*=====================================================================*/

type receiver struct {
	cnt int
}

func (r receiver) Receive() (interface{}, error) {
	i := r.cnt
	r.cnt++
	return i, nil
}

func TestIncoming(t *testing.T) {

	fg := flowgraph.New("test")
	fg.InsertIncoming("incoming", receiver{})
	fg.InsertSink("sink")

	fg.RunAll()
}
