package futil

import (
	"fmt"
	"math/big"
)

func ExampleNewSplitter() {
	s := NewSplitter(2, 64, 67)
	fmt.Printf("%v\n", s)
	fmt.Println("----MaxBits() test-----")
	fmt.Printf("max number of bits: %d\n", s.MaxBits())
	/* Output:
	 */
}
func ExampleSplitter_Split() {
	var parts []*big.Int
	x := big.NewInt(7)
	sp := NewSplitter(2, 5)
	parts = sp.Split(x, parts)
	fmt.Printf("x: %x\n", x)
	fmt.Println("parts:")
	for i := range parts {
		fmt.Printf("%v \n", parts[i])
	}
	/* Output:
	 */
}
func ExampleSplitter_Join() {
	parts := make([]*big.Int, 3)
	parts[0] = big.NewInt(5)
	parts[1] = big.NewInt(10)
	parts[2] = big.NewInt(255)
	x := big.NewInt(2)
	sp := NewSplitter(4, 4, 8)
	err := sp.Join(parts, x)
	fmt.Printf("x: %x\n", x)
	if err != nil {
		fmt.Print(err)
	}
	/* Output:
	 */
}

func ExampleNewSFloatCostValue() {
	Tc := 100.0
	c := NewSFloatCostValue(Tc)
	c.Set(10.0)
	c.Update(15)
	fmt.Printf("c = %v fbits=%f\n", c, c.Fbits())
	/* Output:
	 */
}
