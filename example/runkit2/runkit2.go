/*main gives an example of how to use psokit to run an arbitrary prime factorization problem.
It uses the convenience function simplefactor.Creator() to do so
To run it for say 2 runs type
	go run runkit2.go -nrun 2
at the command line
*/
package main

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso/fun/simplefactor"
	"github.com/mathrgo/setpso/psokit"
)

func main() {
	var p, q, pMin big.Int
	// put in the two prime factors  to test against
	p.SetString("1059652519", 0)
	q.SetString("929636291", 0)
	// choose smallest factor to use
	pMin.SetString("50000000",0)
	// use convenience  creator function
	fc := simplefactor.NewCreator(&p, &q, &pMin)
	// create a run manager
	man := psokit.NewMan()
	// set number of iterations between each data output
	man.SetNthink(120)
	// set the number of particles
	man.SetNpart(61)
	// set run seed linear function of run number
	man.SetPsoSeed(31427, 31)
	// use a built in SPSO optimizer
	man.SetPsoCase("clpso-0")
	// think of a name for the creator
	man.SetFunCase("primeFactoring-1")
	// register the new function creator
	if err := man.AddFun("primeFactoring-1", "attempt to factor a product of primes no 1  ", fc); err != nil {
		fmt.Println(err)
	}
	// choose actions to be run
	if err := man.SelectActs(
		"use-cmd-options",
		"print-headings",
		"print-result",
		"plot-personal-best",
		"run-progress"); err != nil {
		fmt.Println(err)
	}
	// do the runs
	man.Run()
}
