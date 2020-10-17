package psokit

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

/*
SelectActs selects a list of actions by name to be added for used by man. Each Action
is slotted into the runs where they have a capability to act.
*/
func (man *ManPso) SelectActs(ac ...string) error {
	for _, name := range ac {
		var a Act
		switch name {
		case "print-result":
			a = new(Printresult)
		case "print-headings":
			a = new(Printheading)
		case "plot-personal-best":
			a = new(PlotPersonalBest)
		case "use-cmd-options":
			a = new(CmdOptions)
		case "run-progress":
			a = new(RunProgress)
		default:
			a = man.addedAct[name]
			//fmt.Printf("found: %v\n", a)
			if a == nil {
				return fmt.Errorf("Action: %s not found", name)
			}
		}
		//load actions into activity slots to be used while running man
		if ai, ok := a.(ActInit); ok {
			man.actInit = append(man.actInit, ai)
		}
		if ari, ok := a.(ActRunInit); ok {
			//fmt.Printf("RunInit for: %s\n", name)
			man.actRunInit = append(man.actRunInit, ari)
		}
		if au, ok := a.(ActUpdate); ok {
			man.actUpdate = append(man.actUpdate, au)
		}
		if ad, ok := a.(ActData); ok {
			man.actData = append(man.actData, ad)
		}
		if ar, ok := a.(ActResult); ok {
			man.actResult = append(man.actResult, ar)
		}
		if as, ok := a.(ActSummary); ok {
			man.actSummary = append(man.actSummary, as)
		}
	}
	return nil
}

/*
loadActDescription loads the description of the installed
actions
*/
func (man *ManPso) loadActDescription() {
	man.actd = map[string]string{

		"print-result":       "Print final result at end of run; using Printresult ",
		"print-headings":     "Prints setup and run headings; using Printheading",
		"plot-personal-best": "Plots the personal best during a run; using PlotPersonalBest",
		"use-cmd-options":    "Use command options to change configuration; using CmdOptions",
		"run-progress":       "Prints run progress; using RunProgress",
	}
}

/*
ActDescription gives a description of Action by name.
*/
func (man *ManPso) ActDescription() string {
	keys := make([]string, len(man.actd))
	i := 0
	for k := range man.actd {
		keys[i] = k
		i++
	}
	s := fmt.Sprintln("Action Description:")
	sort.Strings(keys)
	for i := range keys {
		k := keys[i]
		s += fmt.Sprintf("%s :\n  %s\n", k, man.actd[k])
	}
	return s
}

/*
Printresult  used as the
Action, print-result
*/
type Printresult struct{}

//Result just prints  the run result as cost and decoded subset
func (a *Printresult) Result(man *ManPso) {
	p := man.P()
	try := p.Part(p.BestParticle()).BestTry()
	fmt.Printf("RUN %d:\n", man.RunID())
	fmt.Printf(" Best Particle: %d\n", p.BestParticle())
	fmt.Printf(" Cost: %s\n", try.Cost())
	fmt.Printf("%s\n", try.Decode())
}

/*
Printheading used as the Action, print-headings
*/
type Printheading struct{}

//Init prints man settings
func (a *Printheading) Init(man *ManPso) {
	fmt.Println(man)
}

/*
RunInit outputs information about the cost-function. If the seed does not
vary between runs it only outputs on the first run
*/
func (a *Printheading) RunInit(man *ManPso) {
	_, sd1 := man.FunSeed()
	if man.RunID() == 0 || sd1 != 0 {
		fmt.Println(man.F().About())
		fmt.Printf("function parameter size in bits %d\n", man.F().MaxLen())
	}
}

// RunProgress output progress statements during the run
type RunProgress struct {
	progress int
}

//RunInit initializes the progres counter
func (a *RunProgress) RunInit(man *ManPso) {
	a.progress = 0
}

//DataUpdate checks  progress and prints out a change in progress
func (a *RunProgress) DataUpdate(man *ManPso) {
	currentProgress := int(10 * man.Diter() / man.Datalength())
	if currentProgress > a.progress {
		a.progress = currentProgress
		if a.progress < 9 {
			fmt.Printf("%d", a.progress)
		} else {
			fmt.Println(a.progress)
		}
	}
}

// ResultsArray is a structure for storing and plotting results from Each
// particle
type ResultsArray struct {
	points []plotter.XYs
}

/*
NewResultsArray creates a Results Array and returns a pointer to it.
It consists of datalength data entries for dimension values per entry.
*/
func NewResultsArray(datalength, dimension int) *ResultsArray {
	var r ResultsArray
	r.points = make([]plotter.XYs, dimension)
	for i := 0; i < dimension; i++ {
		r.points[i] = make(plotter.XYs, datalength)
	}
	return &r
}

/*
ResUpdate puts val into the plotting results array for value index valueID and
data slot dataID where valID is the number of iterations so far in a run.
*/
func (r *ResultsArray) ResUpdate(val float64, dataID, valID, iterID int) {
	r1 := r.points[valID]
	r1[dataID].X = float64(iterID)
	r1[dataID].Y = val
}

//FixLinAxis attempts to give better min/max bounds for linear axis
func FixLinAxis(a *plot.Axis) {

	d := a.Max - a.Min
	if d > 0 {
		scale := math.Pow(10.0, math.Ceil(math.Log10(d)))
		min1 := math.Floor(a.Min/scale) * scale
		rMin := []float64{0.1, 0.2, 0.4, 0.5, 1, 1.2}
		min := min1
		for i := range rMin {
			if s := min1 + scale*rMin[i]; s <= a.Min {
				min = s
			}

		}
		rMax := []float64{2.0, 1.0, 0.5, 0.2, 0.1}
		max := min
		for i := range rMax {
			if s := min + scale*rMax[i]; s >= a.Max {
				max = s
			}
		}
		//fmt.Printf("min=%f max=%f scale=%f min1= %f \n", a.Min, a.Max, scale, min1)
		a.Max = max
		a.Min = min
	}
}

/*
NewPlot creates a basic plot. yaxisname is the Y-axis label; title is the plot title;
runid is the run ID.
*/
func (r *ResultsArray) NewPlot(yaxisname, title string, runid int) {
	// Create a new plot
	pl1, err1 := plot.New()
	if err1 != nil {
		panic(err1)
	}
	// Draw a grid behind the data
	pl1.Add(plotter.NewGrid())
	// for each particle Make a line plotter with points and set its style.
	for i := range r.points {
		pl1Line, _, err := plotter.NewLinePoints(r.points[i])
		if err != nil {
			panic(err)
		}
		pl1.Add(pl1Line)
	}
	pl1.Title.Text = fmt.Sprintf("%s of particle: Run %d", title, runid)
	pl1.X.Label.Text = "iteration"
	pl1.Y.Label.Text = yaxisname
	FixLinAxis(&(pl1.Y))
	// use Log scale on iterations

	//fmt.Printf(" X min = %f \t X max= %f\n", pl1.X.Min, pl1.X.Max)
	//fmt.Printf(" Y min = %f \t Y max= %f\n", pl1.Y.Min, pl1.Y.Max)
	//pl1.X.Min = 10.0
	pl1.X.Scale = plot.LogScale{}
	pl1.X.Tick.Marker = plot.LogTicks{}
	// Save the plot to a PNG file.
	filename := fmt.Sprintf("plot%s%d.pdf", yaxisname, runid)
	if errx1 := pl1.Save(4*vg.Inch, 4*vg.Inch, filename); errx1 != nil {
		panic(errx1)
	}

}

//PlotPersonalBest plots the personal best costs of each  Particle during a run.
// It implements the plot-personal-best Action.
type PlotPersonalBest struct {
	*ResultsArray
}

//RunInit setup plotting arrays for the run
func (pl *PlotPersonalBest) RunInit(man *ManPso) {
	*pl = PlotPersonalBest{NewResultsArray(man.Datalength(), man.Npart())}
}

//DataUpdate loads personal best costs into plot
func (pl *PlotPersonalBest) DataUpdate(man *ManPso) {
	p := man.P()
	for j := 0; j < man.Npart(); j++ {
		pl.ResUpdate(p.Part(j).BestTry().Fbits(), man.Diter(), j, man.Iter())
	}
}

/*
Result generates the plots of the personal best and puts it into  the file of the form:

	plotFbits(Cost)<run ID>.pdf

*/
func (pl *PlotPersonalBest) Result(man *ManPso) {
	pl.NewPlot("Fbits(Cost)", "Personal Best", man.RunID())
}

/*
CmdOptions is the implementation of the Action, use-cmd-options. It provides
the ability to change some of the man options using the command line. if no CmdOptions
are chosen it prints out a list of defaults together with a list of Action names.
*/
type CmdOptions struct{}

//Init reads the command options.
func (cmd *CmdOptions) Init(man *ManPso) {
	var optCase, funCase string
	var debug, listFun, listPso, listAct bool
	var stopAt, nrun, npart int
	flag.StringVar(&optCase, "pso", man.PsoCase(), "name of PSO")
	flag.StringVar(&funCase, "fun", man.FunCase(), "name of function to optimise")
	flag.BoolVar(&debug, "dump", man.DebugDump(), "set to true when debug dumping")
	flag.IntVar(&stopAt, "dbstop", man.StopAt(), "cycle to stop at when doing a debug dump")
	flag.IntVar(&nrun, "nrun", man.Nrun(), "number of independent runs when not debug dumping")
	flag.IntVar(&npart, "npart", man.Npart(), "number of independent runs when not debug dumping")
	flag.BoolVar(&listFun, "listf", false, "list available cost-function")
	flag.BoolVar(&listPso, "listp", false, "list available SPSO")
	flag.BoolVar(&listAct, "lista", false, "list available Actions")

	flag.Parse()
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		fmt.Printf("\n=====================\n %s", man.ActDescription())
		os.Exit(1)
	}
	if err := man.SelectPso(optCase); err != nil {
		fmt.Println(err)
		fmt.Print(man.PsoDescription())
		os.Exit(1)
	}
	if err := man.SelectFun(funCase); err != nil {
		fmt.Println(err)
		fmt.Print(man.FunDescription())
		os.Exit(1)
	}
	man.SetNrun(nrun)
	man.SetNpart(npart)

	if debug {
		man.SetDebugDump(true)
		man.SetStopAt(stopAt)
		man.SetNrun(1)
	} else {
		man.SetDebugDump(false)
	}
	done := false
	if listFun {
		fmt.Println(man.FunDescription())
		done = true
	}
	if listPso {
		fmt.Println(man.PsoDescription())
		done = true
	}
	if listAct {
		fmt.Println(man.ActDescription())
		done = true
	}
	if done {
		os.Exit(0)
	}
}
