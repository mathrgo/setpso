package main

import (
	"fmt"

	"github.com/mathrgo/setpso/psokit"
)

func main() {
	man := psokit.NewMan()
	if err := man.SelectActs(
		"use-cmd-options",
		"print-headings",
		"print-result",
		"plot-personal-best"); err != nil {
		fmt.Println(err)
	}
	man.Run()
}
