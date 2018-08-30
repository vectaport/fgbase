// Package regexp builds regular expressions out of scalable fgbases
package regexp

import (
//	"github.com/vectaport/fgbase"
	"sync/atomic"
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

var CurrID int64 = 0

type Search struct {
	Curr string
	Orig string
	State Mode
	ID int64
}

func NextID() int64 {
        i := atomic.AddInt64(&CurrID, 1)
	if CurrID<0 {
	        panic("possible ID's exceeded")
        }
	return i
}
