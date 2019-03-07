package subsetsum

import (
	"fmt"
	"math/big"
)

func ExampleNew() {
	f := New(8, 10, 3142)
	fmt.Print(f.About())
	fmt.Printf("Cost = %v\n", f.Cost(big.NewInt(0255)))
	pre := big.NewInt(56789)
	hint := big.NewInt(5555)
	fmt.Printf("x= %v\n", pre)
	f.ToConstraint(pre, hint)
	fmt.Printf("constrained x= %v\n", hint)
	//Output:
	//subset value problem parameters:
	//nElements= 8 NBit = 10 Seed= 3142
	//Target value: 1776
	//subset solution:
	//10101101
	//Values:
	//0 	 121
	//1 	 133
	//2 	 377
	//3 	 762
	//4 	 172
	//5 	 158
	//6 	 196
	//7 	 358
	//Cost = 0
	//x= 56789
	//constrained x= 5555
}
