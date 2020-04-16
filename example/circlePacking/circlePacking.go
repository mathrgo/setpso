package main

import (
	"fmt"

	"github.com/mathrgo/setpso/fun/circles"
	"github.com/mathrgo/setpso/psokit"
)

type packFun1 struct{}

func (fc *packFun1) Create(fsd int64) psokit.Fun {
	radius := 0.1
	innerFuzz := 0.1
	outerFuzz := 0.2
	valueNbits := 10
	birthBonus := 1.0
	f := circles.New(radius, innerFuzz, outerFuzz, valueNbits, birthBonus)
	return f
}

func main() {
	var fc packFun1
	skipLen := 1
	dispSize := 500
	ac := circles.NewAnimator(skipLen, dispSize)
	man := psokit.NewMan()
	man.SetNthink(1)
	man.SetNpart(61)
	man.SetPsoCase("clpso-0")
	man.SetFunCase("circle-1")
	if err := man.AddFun("circle-1", "attempts to pack circles into a unit circle ", &fc); err != nil {
		fmt.Println(err)
	}
	if err := man.AddAct("animate-1", " animates used circles of global best", ac); err != nil {
		fmt.Println(err)
	}
	if err := man.SelectActs(
		"use-cmd-options",
		"print-headings",
		"print-result",
		"plot-personal-best",
		"run-progress",
		"animate-1"); err != nil {
		fmt.Println(err)
	}
	man.Run()

}
