package flowgraph

import (
	"fmt"
	"math"
	"testing"
	"time"
	"reflect"
)

func promoteTest(a, b, x Edge) {
	
	for {
		aTmp := <- (*a.Data)[0]
		bTmp := <- (*b.Data)[0]
		fmt.Printf("%v,%v --> ", reflect.TypeOf(aTmp), reflect.TypeOf(bTmp))
		
		aBig,bBig,same := Promote(nil, aTmp, bTmp)
		
		fmt.Printf("%v,%v,%v\n", reflect.TypeOf(aBig), reflect.TypeOf(bBig), same);
		
		(*x.Data)[0] <- aBig
	}
	
	
}
func TestPromote(t *testing.T) {

	a := MakeEdge("a",nil)
	var ad = (make([]chan Datum,1))
	a.Data = &ad
	(*a.Data)[0] = make(chan Datum)
	b := MakeEdge("b",nil)
	var bd = make([]chan Datum,1)
	b.Data = &bd
	(*b.Data)[0] = make(chan Datum)
	x := MakeEdge("x",nil)
	var xd = make([]chan Datum,1)
	x.Data = &xd
	(*x.Data)[0] = make(chan Datum)

	go promoteTest(a, b, x)

  	var answer interface {}
	(*a.Data)[0] <- 512
	(*b.Data)[0] <- int8(4)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	(*a.Data)[0] <- byte(4)
	(*b.Data)[0] <- 512
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	(*a.Data)[0] <- byte(4)
	(*b.Data)[0] <- byte(100)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))
	
	(*a.Data)[0] <- "abcdef"
	(*b.Data)[0] <- byte(4)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- int8(8)
	(*b.Data)[0] <- uint32(4)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- 1 + 0i
	(*b.Data)[0] <- uint32(4)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- complex(float32(1),0)
	(*b.Data)[0] <- int64(4)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- float32(0)
	(*b.Data)[0] <- byte(0)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- uint64(math.MaxUint64)
	(*b.Data)[0] <- int64(-1)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- uint64(math.MaxUint64>>2)
	(*b.Data)[0] <- int64(-1)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	(*a.Data)[0] <- rune(33)
	(*b.Data)[0] <- int8(-1)
        answer = <- (*x.Data)[0]
	fmt.Printf("answer is %v of type %v\n\n", answer, reflect.TypeOf(answer))

	time.Sleep(time.Second)

}

