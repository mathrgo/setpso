package multimode

import (
	"fmt"
	"math/big"
	"math/rand"
)

func streamFbits(n int) []float64 {
	nmode := 4
	nbits := 16
	margin := 1.0
	sigma := 0.1
	Tc := 100.0
	sigmaThres := 2.0
	f := NewFun(nmode, nbits, margin, sigma,
		Tc, sigmaThres, 3142)
	cost := f.NewCostValue()

	z := new(big.Int)
	z.SetInt64(1000)
	out := make([]float64, n)
	for i := range out {
		cost.Update(f.Cost(z))
		out[i] = cost.Fbits()
		rand.Intn(6)
	}
	return out
}

func ExampleNewFun() {
	n := 10000001
	s1 := streamFbits(n)
	s2 := streamFbits(n)
	count := 0
	for i := range s1 {
		if s1[i] != s2[i] {
			count++
		}

	}
	fmt.Printf("mismatch count = %d", count)
	/* Output:
	mismatch count = 0
	*/
}
