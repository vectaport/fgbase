// Package regexp builds regular expressions out of scalable flowgraphs
package regexp

import (
//	"github.com/vectaport/flowgraph"
)

type Mode int
const (
	Done Mode = iota
	Live
	Fail
)

var Modes = map[Mode]string {
	Done: "Done",
	Live: "Live",
	Fail: "Fail",
}

func (m Mode) String() string {
	return Modes[m]
}

type Search struct {
	Curr string
	Orig string
	State Mode
}
