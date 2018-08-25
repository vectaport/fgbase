package flowgraph_test

import (
        "github.com/vectaport/flowgraph"
	"testing"
)

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

func TestAddInput(t *testing.T) {
	// Different allocations should not be equal.
	fg := flowgraph.New("abc")
	fg.AddInput("input1")
	if fg.InputName(0)!="input1" {
		t.Errorf(`AddInput followed by InputName doesn't work`)
	}
}

func TestAddOutput(t *testing.T) {
	// Different allocations should not be equal.
	fg := flowgraph.New("abc")
	fg.AddOutput("output1")
	if fg.OutputName(0)!="output1" {
		t.Errorf(`AddOutput followed by OutputName doesn't work`)
	}
}





       