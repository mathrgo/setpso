/*
main gives an example of a multimode function with added noise for testing the SPSO using the futil.SFloatCostvalue CostValue interface
*/
package main

import (
	"fmt"

	"github.com/mathrgo/setpso/fun/multimode"
	"github.com/mathrgo/setpso/psokit"
)

//NewFun1 is used to create example 1 of the function
type NewFun1 struct {
}

//Create creates the example 1 for PSO kit.
func (fc *NewFun1) Create(fsd int64) psokit.Fun {
	nmode := 4
	nbits := 20
	margin := 1.0
	sigma := 0.1
	Tc := 10000.0
	sigmaThres := 4.0

	f := multimode.NewFun(nmode, nbits, margin, sigma,
		Tc, sigmaThres, fsd)
	return f
}

func main() {
	var fc NewFun1
	man := psokit.NewMan()
	man.SetNthink(60)
	man.SetNpart(3)
	man.SetPsoCase("clpso-0")
	man.SetFunCase("mms-1")
	if err := man.AddFun("mms-1", "multimode function with noise", &fc); err != nil {
		fmt.Println(err)
	}
	if err := man.SelectActs(
		"use-cmd-options",
		"print-headings",
		"print-result",
		"plot-personal-best",
		"run-progress"); err != nil {
		fmt.Println(err)
	}
	man.Run()

}
