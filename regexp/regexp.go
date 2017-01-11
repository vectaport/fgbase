// Package regexp builds regular expressions out of scalable flowgraphs
package regexp

import (
//	"github.com/vectaport/flowgraph"
)

type Mode int
const (
	Live Mode = iota
	Done
	Fail
)

var Modes = map[Mode]string {
	Live: "Live",
	Done: "Done",
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
