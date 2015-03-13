package main

import (
	"github.com/vectaport/flowgraph"
	"fmt"
	"math"
	"time"
	"reflect"
)

func promote_test(a, b, x flowgraph.Conn) {
	
	for {
		_a := <- a.Data
		_b := <- b.Data
		fmt.Printf("%v,%v --> ", reflect.TypeOf(_a), reflect.TypeOf(_b))
		
		_abig,_bbig := flowgraph.Promote(_a, _b)
		
		fmt.Printf("%v,%v\n", reflect.TypeOf(_abig), reflect.TypeOf(_bbig));
		
		x.Data <- _abig
	}
	
	
}
func main() {

	a := flowgraph.MakeConn(false,true,nil)
	b := flowgraph.MakeConn(false,true,nil)
	x := flowgraph.MakeConn(false,true,nil)

	go promote_test(a, b, x)

  	var answer interface {}
	a.Data <- 512
	b.Data <- int8(4)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	a.Data <- byte(4)
	b.Data <- 512
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	a.Data <- byte(4)
	b.Data <- byte(100)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	a.Data <- "abcdef"
	b.Data <- byte(4)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- int8(8)
	b.Data <- uint32(4)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- 1 + 0i
	b.Data <- uint32(4)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- complex(float32(1),0)
	b.Data <- int64(4)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- float32(0)
	b.Data <- byte(0)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- uint64(math.MaxUint64)
	b.Data <- int64(-1)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- uint64(math.MaxUint64>>2)
	b.Data <- int64(-1)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	a.Data <- rune(33)
	b.Data <- int8(-1)
        answer = <- x.Data
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	time.Sleep(1000000000)

}

