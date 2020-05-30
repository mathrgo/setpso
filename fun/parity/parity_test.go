package parity

import (
	"fmt"
	"math/big"
	"math/rand"
)

func ExampleNewSampler() {
	p := NewSampler(4)
	x := big.NewInt(0)
	y := big.NewInt(0)
	fmt.Printf("About:\n%s\n", p.About())
	rnd := rand.New(rand.NewSource(3142))
	for i := 0; i < 16; i++ {
		p.Sample(x, y, rnd)
		fmt.Printf("%b => %b\n", x, y)
	}

	//Output:
}
