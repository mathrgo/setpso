/*
main gives an example of how to find  DAG modeling of parity checking using random samples.
it uses the provided methods in fun/dag subdirectory.
To run it for 2 runs type
	go run parityDAG.go -nrun 2
at the command line
*/
package main

import (
	"fmt"

	"math/rand"
	"github.com/mathrgo/setpso/fun/dag"
	"github.com/mathrgo/setpso/fun/parity"
	"github.com/mathrgo/setpso/psokit"
)

type parityFun1 struct {
}

func (fc *parityFun1) Create(fsd int64) psokit.Fun {
	s := parity.NewSampler(5)
	opt := dag.NewOpt4Bool()
	nnode := 6
	nbitslookback := 3
	sizeCostFactor := int64(10)
	sampleSize := 32
	rnd := rand.New(rand.NewSource(fsd))
	f := dag.NewFunBool(nnode, nbitslookback, opt, sizeCostFactor,
		s, sampleSize, rnd)
	return f
}

func main() {

	var fc parityFun1
	man := psokit.NewMan()
	man.SetNthink(10)
	man.SetNpart(61)
	man.SetPsoCase("clpso-0")
	man.SetFunCase("parity-5-1")
	if err := man.AddFun("parity-5-1", "attempts to find 4bool DAG for parity of 5 inputs ", &fc); err != nil {
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
