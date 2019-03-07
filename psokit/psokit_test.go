package psokit

import (
	"fmt"

	"github.com/mathrgo/setpso/fun/subsetsum"
)

type myFun struct{}

func (fc *myFun) Create(fsd int64) Fun {
	return subsetsum.New(50, 10, fsd)
}

func ExampleNewMan() {
	var fc myFun
	man := NewMan()
	// try adding creator to existing cost-function
	if err := man.AddFun("subsetsum-0", "subsetsum of 50 elements", &fc); err != nil {
		fmt.Println(err)
	}
	// try selecting a non existent Fun
	if err := man.SelectFun("subsetsum-1"); err != nil {
		fmt.Println(err)
	}
	//try deleting a cost-function that has not been added
	if err := man.DelFun("subsetsum-0"); err != nil {
		fmt.Println(err)
	}
	// this should be the default
	fmt.Println("\n==default man==")
	fmt.Print(man)
	// add a new  cost function
	if err := man.AddFun("subsetsum-1", "subsetsum of 50 elements", &fc); err != nil {
		fmt.Println(err)
	}
	// select it
	if err := man.SelectFun("subsetsum-1"); err != nil {
		fmt.Println(err)
	}
	// it is now managed by man
	fmt.Println("\n===man with new cost-function==")
	fmt.Print(man)
	fmt.Print(man.FunDescription())

	//delete a function that is added and selected for the runs
	if err := man.DelFun("subsetsum-1"); err != nil {
		fmt.Println(err)
	}
	// it is now not managed by man which uses the default
	fmt.Println("\n===man with default cost-function==")
	fmt.Print(man)
	fmt.Print(man.PsoDescription())
	/* Output:
	 */
}

func ExampleManPso_SelectActs() {
	man := NewMan()
	if err := man.SelectActs("print-result"); err != nil {
		fmt.Println(err)
	}
	// ensure there are instances to apply print-result
	man.Init()
	// go under the hood to see if SelectActs() has installed the action
	r := man.actResult[0]
	r.Result(man)
	/* Output:
	 */
}
