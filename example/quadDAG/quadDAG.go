/*
main gives an example of how to find  DAG modeling of quadratic equation solution.
it uses the provided methods in fun/dag subdirectory.
To run it for 1 run type
	go run quadDAG.go -nrun 1
at the command line
*/
package main

import (
	"fmt"

	"math/rand"

	"github.com/mathrgo/setpso/fun/dag"
	"github.com/mathrgo/setpso/fun/quadratic"
	"github.com/mathrgo/setpso/psokit"
)

type quadFun1 struct {
}

func (fc *quadFun1) Create(fsd int64) psokit.Fun {
	s := quadratic.NewExSampler(10.0)
	C := dag.NewInt2FloatRange(16, 0.0, 2.0)
	P := dag.NewInt2FloatList(2.0, 0.5, 1.0, 0.5)
	opt := dag.NewOptMorphFloat(C, P)
	nnode := 3
	nbitslookback := 2
	sizeCostFactor := 0.1
	sampleSize := 10000
	Tc := 100.0
	sigThres := 4.0
	bufLife := 100000
	rnd := rand.New(rand.NewSource(fsd))
	f := dag.NewFunFloat(nnode, nbitslookback, opt, sizeCostFactor,
		s, sampleSize, rnd, Tc, sigThres, bufLife)
	return f
}

func main() {

	var fc quadFun1
	man := psokit.NewMan()
	man.SetNthink(1)
	man.SetNpart(10)
	man.SetPsoCase("clpso-0")
	man.SetFunCase("quad-10-1")
	if err := man.AddFun("quad-10-1", "attempts to find quadratic solution formula ", &fc); err != nil {
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
