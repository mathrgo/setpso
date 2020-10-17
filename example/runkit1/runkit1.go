package main

import (
	"fmt"

	"github.com/mathrgo/setpso/psokit"
)

func main() {
	man := psokit.NewMan()
	// set noise seed to fourth run values to give zero cost result
	man.SetPsoSeed(578+3*34,34)
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
