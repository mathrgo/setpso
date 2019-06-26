//this runs the sum of 3 cubes for 42 finder
package main

import (
	"fmt"

	"github.com/mathrgo/setpso/fun/cubes3442"
	"github.com/mathrgo/setpso/psokit"
)

type sc42Fun1 struct{}

func (fc *sc42Fun1) Create(fsd int64) psokit.Fun {
	return cubes3442.New(60)
}
func main() {
	var fc sc42Fun1
	man := psokit.NewMan()
	man.SetNthink(1200)
	man.SetNpart(61)
	man.SetPsoCase("clpso-0")
	man.SetFunCase("sumcubesfor42-1")
	if err := man.AddFun("sumcubesfor42-1", "attempts to find sum of 3 cubes for 42 ", &fc); err != nil {
		fmt.Println(err)
	}
	if err := man.SelectActs(
		"use-cmd-options",
		"print-headings",
		"print-result",
		"plot-personal-best"); err != nil {
		fmt.Println(err)
	}
	man.Run()
}
