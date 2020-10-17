package multimode

import (
	"fmt"
	"math/big"
)

func ExampleNewFun() {
	nMode := 2
	nbits := 16
	margin := 1.0
	sigma := 0.0
	Tc := 100.0
	SigmaMargin:=2.0

	f := NewFun(nMode, nbits, margin, sigma,
		Tc,SigmaMargin, 3142)
	fmt.Print(f.About())

	z := new(big.Int)
	z.SetInt64(9811)
	t := f.NewTry()
	f.SetTry(t, z)
	for i:=0;i<200; i++{
		f.UpdateCost(t)
	}
	fmt.Printf(" %s \n", t.Decode())
	fmt.Printf("Param= %v\n", t.Parameter())
	fmt.Printf(" %s",t.Cost())

	/* Output:
	 */
}
