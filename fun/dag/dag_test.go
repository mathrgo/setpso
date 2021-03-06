package dag

import (
	"fmt"
	"math/big"
	"math/rand"
	"github.com/mathrgo/setpso/fun/parity"
)

func ExampleNewOpt4Bool() {
	o := NewOpt4Bool()
	opt := uint(6) // code for exclusive or
	fmt.Printf("Encoding bit size: %d\n", o.BitSize())
	fmt.Printf("Symbol:[%s]\n", o.Decode(opt))
	fmt.Printf("Node cost: %d\n", o.Cost(opt))
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			fmt.Printf("l=%d r= %d => %d\n",
				i, j, o.Opt(uint(i), uint(j), opt))
		}

	}

	// Output:
	// Encoding bit size: 4
	// Symbol:[ + ]
	// Node cost: 1
	// l=0 r= 0 => 0
	// l=0 r= 1 => 1
	// l=1 r= 0 => 1
	// l=1 r= 1 => 0
}
func ExampleNewDag() {
	opt := NewOpt4Bool()
	nvar := 4
	nnode := 4
	nout := 1
	nbitslookback := 4

	d := NewDag(nvar, nnode, nout, nbitslookback, opt)
	z := big.NewInt(0)
	fmt.Sscanf("056036016", "%x", z)
	fmt.Printf("z: %b\n", z)
	t:= d.CreateData()
	d.IDecode(t,z)
	fmt.Printf("Number of bits to encode: %d\n", d.MaxLen())
	fmt.Printf("Dag Decode:\n %s\n", t.Decode())
	//Output:
}

func ExampleNewFunBool() {
	s := parity.NewSampler(4)
	opt := NewOpt4Bool()
	nnode := 4
	nbitslookback := 4
	sizeCostFactor := int64(1)
	sampleSize := 16
	rnd := rand.New(rand.NewSource(3142))

	f := NewFunBool(nnode, nbitslookback, opt, sizeCostFactor,
		s, sampleSize, rnd)
	fmt.Printf("About:\n %s\n", f.About())
	z := big.NewInt(0)
	fmt.Sscanf("056038017", "%x", z)
	fmt.Printf("z: %b\n", z)
	t:= f.CreateData()
	f.IDecode(t,z)
	fmt.Printf("Number of bits to encode: %d\n", f.MaxLen())
	fmt.Printf("Dag Decode:\n %s\n", t.Decode())
	cost:=big.NewInt(0)
	f.Cost(t,cost)
	fmt.Printf("Cost: %v\n", cost)
	//Output:
}

