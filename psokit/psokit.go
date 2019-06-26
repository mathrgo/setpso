package psokit

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso"
)

// DefaultFun is the default cost function.
const DefaultFun = "subsetsum-0"

// DefaultPso is the default SPSO.
const DefaultPso = "gpso-0"

//PsoInterface  is the PsoInterface from setpso.
type PsoInterface = setpso.PsoInterface

//Fun is the cost function interface from setpso.
type Fun = setpso.Fun

//CreatePso is the interface for creating new instances of SPSO.
type CreatePso interface {
	Create(p0 *setpso.Pso, hu ...*setpso.PsoHeuristics) PsoInterface
}

//CreateFun is the interface for  creating new instances of cost-functions.
type CreateFun interface{ Create(sd int64) Fun }

/*
CreateAct is the interface for  creating new instances of Action. each action
must have the appropriate methods to indicate where it will be used in a run
sequence. the supported interfaces are:
		ActInit // pre runs Action
		ActRunInit // pre run Action
		ActUpdate  // post iteration Action
		ActData  // data interface Action occurring every Nthink() iterations
		ActResult // post run Action
		ActSummary // post runs Action
*/
type CreateAct interface{ Create(sd int64) Act }

/*
ActInit is the interface for pre runs initialising Action. This is expected to
be run before the cost-function and SPSO instances are available; it is used to
configure man before the run commences; it typically is used to provide a
command line interface or even include new things that are not part of the
installed options.
*/
type ActInit interface{ Init(man *ManPso) }

/*
ActRunInit is the interface for a run initialising Action. It can change the
cost-function and SPSO instances and even swap them if you are not interested in
the runs being statistically independent!
*/
type ActRunInit interface{ RunInit(man *ManPso) }

/*
ActUpdate is the interface for post Update  per iteration Action. This may be
used for ultra fine  logging of  data during a debug dump, update a variance
calculation, stop the run before it reaches a maximum number of iterations or
even change the nature of the function being optimised.
*/
type ActUpdate interface{ Update(man *ManPso) }

/*
ActData  is the interface for data input/output actions that occur at fixed intervals
of NThink() iterations. This is used to reduce the communication bandwidth for memory demanding
actions such as plotting the progress of the swarm.
*/
type ActData interface{ DataUpdate(man *ManPso) }

//ActResult is the interface for post run Action.
type ActResult interface{ Result(man *ManPso) }

//ActSummary is the  interface for post runs Action.
type ActSummary interface{ Summary(man *ManPso) }

/*
Act is  for arbitrary Action, which is placed in the appropriate slots based on
its methods. */
type Act interface{}

// ManPso manages the runs.
type ManPso struct {
	f Fun
	// current cost function instance name
	funCase string
	// cost functions descriptions
	fund map[string]string
	// pointers to added cost function instance creator
	addedFun map[string]CreateFun

	p PsoInterface
	// current SPSO instance name
	psoCase string
	// SPSO instance  descriptions
	psod map[string]string
	// pointers to added SPSO instance creator
	addedPso map[string]CreatePso

	// description of Actions by name
	actd map[string]string
	// pointers to added action creator
	addedAct map[string]CreateAct
	// actions by point of application
	actInit    []ActInit
	actRunInit []ActRunInit
	actUpdate  []ActUpdate
	actData    []ActData
	actResult  []ActResult
	actSummary []ActSummary

	//iteration count during run
	iter int
	//data count
	diter int
	// maximum number of data outputs per run
	ndata int
	// thinking iteration between data output
	titer int
	// number of thinking iterations between data output
	nthink int

	// run id
	runid int
	// iteration number to stop at when debug = true
	stopAt int
	// true when detailed output to file is used
	dbug bool
	// number of runs
	nrun int
	// number of particles in SPSO instance
	npart int
	// cost function seed gain against run id
	funSeed1 int64
	// cost function seed offset
	funSeed0 int64
	// SPSO seed gain against run id
	psoSeed1 int64
	// SPSO seed offset
	psoSeed0 int64
}

/*
NewMan creates a default instance of ManPso.
*/
func NewMan() *ManPso {
	man := new(ManPso)
	man.funCase = "subsetsum-0"
	man.fund = make(map[string]string)
	man.addedFun = make(map[string]CreateFun)
	man.loadFunDescription()
	man.psoCase = DefaultPso
	man.psod = make(map[string]string)
	man.addedPso = make(map[string]CreatePso)
	man.loadPsoDescription()
	man.actd = make(map[string]string)
	man.addedAct = make(map[string]CreateAct)
	man.actInit = make([]ActInit, 0, 10)
	man.actRunInit = make([]ActRunInit, 0, 10)
	man.actUpdate = make([]ActUpdate, 0, 10)
	man.actData = make([]ActData, 0, 10)
	man.actResult = make([]ActResult, 0, 10)
	man.actSummary = make([]ActSummary, 0, 10)
	man.loadActDescription()
	man.stopAt = 3
	man.dbug = false
	man.ndata = 1200
	man.nthink = 300
	man.nrun = 1
	man.npart = 10
	man.funSeed1 = 0
	man.funSeed0 = 3142
	man.psoSeed1 = 34
	man.psoSeed0 = 578
	return man
}

/*
Init sets up man with instance of cost-function and SPSO based on its settings.
This is automatically  called  at the beginning of each run so is normally not
explicitly called; it is  exportable for those that need it. For instance it is
used in some test examples where there is no need to run a case.
*/
func (man *ManPso) Init() {
	man.CreateFun(man.funCase)
	man.CreatePso(man.psoCase)
}

//F returns the instance of the cost-function in use for Actions during a run
func (man *ManPso) F() Fun { return man.f }

//P returns the instance of the SPSO in use for Actions during a run
func (man *ManPso) P() PsoInterface { return man.p }

/*
String gives a description of the man settings
*/
func (man *ManPso) String() string {
	s := "ManPso Settings:\n"
	s += fmt.Sprintf("cost-function = %s\t", man.funCase)
	s += fmt.Sprintf("SPSO = %s\n", man.psoCase)
	if man.dbug {
		s += fmt.Sprintf("Detailed debug for %d iterations\n", man.stopAt)
	}
	s += fmt.Sprintf("Number of Runs = %d \t", man.nrun)
	s += fmt.Sprintf("Number of Particles = %d\n", man.npart)
	s += fmt.Sprintf("Max number of data coms in a run = %d\n", man.ndata)
	s += fmt.Sprintf("Thinking interval between data coms = %d\n", man.nthink)
	s += fmt.Sprintf("funSeed=%d + runid*%d\t", man.funSeed0, man.funSeed1)
	s += fmt.Sprintf("psoSeed=%d + runid*%d\n", man.psoSeed0, man.psoSeed1)
	return s
}

//FunCase returns the cost-function case name.
func (man *ManPso) FunCase() string { return man.funCase }

//SetFunCase sets the cost-function case name.
func (man *ManPso) SetFunCase(name string) {
	man.funCase = name
}

//PsoCase returns the SPSO case name.
func (man *ManPso) PsoCase() string { return man.psoCase }

//SetPsoCase sets the SPSo case name.
func (man *ManPso) SetPsoCase(name string) {
	man.psoCase = name
}

//Iter returns the iteration count during a run.
func (man *ManPso) Iter() int { return man.iter }

//Diter returns the data output count during a run.
func (man *ManPso) Diter() int { return man.diter }

//SetNdata sets Ndata.
func (man *ManPso) SetNdata(n int) { man.ndata = n }

//Ndata is the maximum number of data events per run used mainly for plotting.
func (man *ManPso) Ndata() int { return man.ndata }

//Nthink is the number of thinking iterations between data output.
func (man *ManPso) Nthink() int { return man.nthink }

//SetNthink sets Nthink.
func (man *ManPso) SetNthink(n int) { man.nthink = n }

//RunID returns run Number.
func (man *ManPso) RunID() int { return man.runid }

//SetStopAt sets StopAt.
func (man *ManPso) SetStopAt(n int) { man.stopAt = n }

//StopAt returns iteration number to stop at when debug = true.
func (man *ManPso) StopAt() int { return man.stopAt }

//SetDebugDump sets DebugDump flag.
func (man *ManPso) SetDebugDump(db bool) { man.dbug = db }

//DebugDump returns true when detailed output to file is used.
func (man *ManPso) DebugDump() bool { return man.dbug }

//SetNrun sets Nrun.
func (man *ManPso) SetNrun(n int) { man.nrun = n }

//Nrun returns number of runs.
func (man *ManPso) Nrun() int { return man.nrun }

//SetNpart sets Npart.
func (man *ManPso) SetNpart(n int) { man.npart = n }

//Npart returns number of particles in SPSO instance.
func (man *ManPso) Npart() int { return man.npart }

/*
PsoSeed returns the random generator seed components of SPSO
where seed=sd0+sd1*RunId().
*/
func (man *ManPso) PsoSeed() (sd0, sd1 int64) {
	return man.psoSeed0, man.psoSeed1
}

//SetPsoSeed sets the SPSO seed components.
func (man *ManPso) SetPsoSeed(sd0, sd1 int64) {
	man.psoSeed0 = sd0
	man.psoSeed1 = sd1
}

/*
FunSeed returns the random generator seed components of the cost-function
where seed=sd0+sd1*RunId()
*/
func (man *ManPso) FunSeed() (sd0, sd1 int64) {
	return man.funSeed0, man.funSeed1
}

//SetFunSeed sets the cost_function seed components.
func (man *ManPso) SetFunSeed(sd0, sd1 int64) {
	man.funSeed0 = sd0
	man.funSeed1 = sd1
}

/*
Run runs the chosen SPSO using the chosen cost-function and settings in man for
Nrun() runs. Each run starts with a new cost-function and SPSO with different
but computed random number seeds to aim at making each run independent of other
runs in the sequence thus making it easy to generate  moderately unbiased
performance statistics.

During the runs chosen Actions are activated according to their interfaces.
*/
func (man *ManPso) Run() {
	for i := range man.actInit {
		man.actInit[i].Init(man)
	}
	for man.runid = 0; man.runid < man.nrun; man.runid++ {
		man.iter = 0
		man.Init()
		for i := range man.actRunInit {
			man.actRunInit[i].RunInit(man)
		}
		for man.diter = 0; man.diter < man.ndata; man.diter++ {
			for man.titer = 0; man.titer < man.nthink; man.titer++ {
				man.p.Update()
				for i := range man.actUpdate {
					man.actUpdate[i].Update(man)
				}
				man.iter++
			}
			for i := range man.actData {
				man.actData[i].DataUpdate(man)
			}
		}
		for i := range man.actResult {
			man.actResult[i].Result(man)
		}
	}
	for i := range man.actSummary {
		man.actSummary[i].Summary(man)
	}
}

/*
AddFun adds a cost function instance f with an assigned name to reference it by
where desc is the description of the function. Note one cannot here reuse
instance names. It returns true if successful. However, if there is a need to
reuse a name that has been added, call DelFun() to remove it.
*/
func (man *ManPso) AddFun(name, desc string, f CreateFun) error {
	if man.fund[name] == "" {
		man.fund[name] = desc
		man.addedFun[name] = f
		return nil
	}
	return fmt.Errorf("attempted to add %s to a cost-function creator that exists ", name)

}

/*
DelFun can be used to delete an added cost function thus freeing resources. it
logs the event if not found. As a temporary measure man is set to use the
Default cost function if name has been selected.
*/
func (man *ManPso) DelFun(name string) error {
	if man.addedFun[name] == nil {
		return fmt.Errorf("Could not delete  cost-function creator %s", name)
	}
	delete(man.addedFun, name)
	delete(man.fund, name)
	if name == man.funCase {
		man.funCase = DefaultFun
	}
	return nil
}

/*
AddPso adds a SPSO instance creator p with an assigned name to reference it by
where desc is the description of it. Note one cannot here reuse instance names.
However, if there is a need to reuse a name that has been added, call DelFun()
to remove it.
*/
func (man *ManPso) AddPso(name, desc string, p CreatePso) error {
	if man.psod[name] == "" {
		man.psod[name] = desc
		man.addedPso[name] = p
		return nil
	}
	return fmt.Errorf("attempted to add %s to a SPSO creator that exists ", name)

}

/*
DelPso can be used to delete an added SPSO instance creator thus freeing resources.
*/
func (man *ManPso) DelPso(name string) error {
	if man.addedPso[name] == nil {
		return fmt.Errorf("Could not delete  SPSO creator %s", name)
	}
	delete(man.addedPso, name)
	delete(man.psod, name)
	if name == man.psoCase {
		man.psoCase = DefaultPso
	}
	return nil

}

/*
AddAct adds an Action instance creator to man for it to use later on.
the Action has the name name and description desc. If the Action already exists
it is not added and an error message is returned.
*/
func (man *ManPso) AddAct(name, desc string, a CreateAct) error {
	if man.actd[name] == "" {
		man.actd[name] = desc
		man.addedAct[name] = a
		return nil
	}
	return fmt.Errorf("attempted to add %s to a Action creator that exists ", name)
}

/*
Fbits gives a floating point measure of number of bits  in x that takes on non
integer values to help represent big integer size for plotting. it approximates
to the log of the big integer.
*/
func Fbits(x *big.Int) float64 {
	n := x.BitLen()
	if n <= 0 {
		return float64(0)
	}
	n--
	var a big.Int
	a.SetBit(&a, n, 1)
	var r big.Rat
	r.SetFrac(x, &a)
	f, _ := r.Float64()
	return f + float64(n)

}

/*
Fdif gives a measure of the difference of the integers x,y as sets and returns
the result as a floating point value of the Hamming distance between x and y.
*/
func Fdif(x, y *big.Int) float64 {
	var z big.Int
	z.Xor(x, y)
	return float64(setpso.CardinalSize(&z))
}
