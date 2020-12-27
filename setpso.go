package setpso

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"

	"github.com/mathrgo/setpso/fun/futil"
)

//Try interface from futil
// It essentially contains all the data that depends on the parameter being tried for a solution of the optimization problem.
type Try = futil.Try

/*
Fun is the interface to the function that evaluates the cost of tries. The try parameter is a binary pattern stored as a big
integer. The integer can be regarded as a sub set of integers from 0 to
MaxLen()-1 where i in this range is a member of the sub set if the ith bit is
1.  MaxLen() returns  the maximum number of bits in
the pattern. ToConstraint() attempts to modify hint to be constraint
satisfying. while possibly using pre  to aid in this process. pre should not be
changed during this process. It returns true if it succeeds. By convention pre
always corresponds to a Parameter (apart from when it is the empty set at the
beginning) that was a personal best and thus satisfies the constraints required
by the function to give a meaningful solution. The function should be able to
find a constraint satisfying version for  hint that approximates to hint  most
of the time so that the SPSO can frequently change the Particle's Parameter
during an update.

Providing a Meaningful Interpretation of Parameters

Representing a particle's Parameters as a binary string or subset may not be
easily  interpreted  as a meaningful representation of a solution to a
combinatorial optimization so the cost function must provide an easily
understood description of a Parameter as a string supplied by the Decode()
function.

Catering for Changing a  Set Item

This is an advanced feature for cost functions. Certain set items will after a
while not be used to give personal best at which point a cost function can be
hinted to remove this item through Delete() which returns true if the cost
function does so. if the cost function  returns true it is then up to the cost
function to either replace this by a new item with its own contribution to cost
or set the  corresponding item in a hint to zero during a successful
ToConstraint() call. in this way the cost function  can try out alternative
items that may result in an improved cost. Typically this feature is not used
and a call to  Delete() returns false.
*/
type Fun interface {
	//creates a new Try which is a pointer to a structure using  a default parameter that should satisfy constraints and updates Try's cost and internal decode.
	NewTry() Try
	// setup t from x, where z is parameter assumed to satisfy constraints and it also computes dependants such as cost
	SetTry(t Try, z *big.Int)

	// sets  the try by copying src to dest
	Copy(dest, src Try)

	// updates cost evaluation of a try where a lower cost is better
	// this is needed when the cost changes
	UpdateCost(x Try)

	// Cmp checks to see if y is better than x
	// for mode = futil.CostMode:
	// compare two tries with a spread of uncertainty from -1.0 to 1.0. -1.0 for
	// definitely cost(x) < cost(y) and 1.0 for definitely cost(x) > cost(y).
	// A deterministic cost should always return -1.0 or 1.0 with a value of
	// -1.0 if the costs are equal.
	// for mode = futil.TriesMode: (used only when cost is not deterministic)
	// compare how successful  y has been in being better than  an x
	//and updates y's success stats.
	// While being compared a value > 1.0 indicates y should replace x; a value
	// < -1.0 indicates y should be removed as a candidate.
	Cmp(x, y Try, mode futil.CmpMode) float64

	// maximum number of bits in the parameter big integer which is the
	// maximum number of elements in the subset
	// at the moment this is fixed during a run but in the future could vary possibly to incorporate some form of subroutine use.
	MaxLen() (maxlen int)
	// string giving a description of the cost function
	About() (s string)
	// this attempts to  give a constraint satisfying Try  in  pre that matches the hint by modifying hint
	// pre is the previous try to be replaced. on success it returns true otherwise it returns false and should leave pre un changed.
	ToConstraint(pre Try, hint *big.Int) bool

	// this hints to the function to replace the ith particle-item
	// if it returns True the function has replaced the item with a new meaning
	// thus modifying the decoder.
	Delete(i int) bool
}

/*
PsoInterface is the interface used with the PSO  and its variants. Normally
Pso supplies this apart from the Update() function. The interface is used by
the monitoring and running program  and psokit uses this interface to provide
useful plotting functions.
*/
type PsoInterface interface {
	// single state update to all particles in the PSO
	Update()
	// best particle Id
	BestParticle() int
	// number of particles
	Nparticles() int
	// array of particles
	Part(i int) *Particle
	// current try of the ith Particle
	CurrentTry(i int) Try
	// Local-best try of the ith Particle
	LocalBestTry(i int) Try
	// interface for requesting debug output based on cmd
	PrintDebug(w io.Writer, cmd string)
	//Heuristics returns a copy of the master heuristics
	Heuristics() *PsoHeuristics
	//SetHeuristics sets the heuristics for the particle swarm.
	SetHeuristics(hu *PsoHeuristics)
}

// Particle is the state of a member of the PSO
type Particle struct {
	// current try
	current Try
	// list of tries that did not quite make it to bestTry
	tries []Try
	// this is used as a raw update before mapping to a constraint satisfying
	// state
	hint *big.Int
	// pointer to group the particle belongs to
	group *Group
	// best try so far
	bestTry Try
	// current  flipping probability requests for each bit component
	vel []float64

	debug bool
}

//BestTry returns the best try
func (p *Particle) BestTry() Try {
	return p.bestTry
}

//CurrentTry returns current try
func (p *Particle) CurrentTry() Try {
	return p.current
}

//Group is a collection of particles with same heuristic settings
type Group struct {
	//group's id
	id string
	// array of member particles
	members []int
	// list of update targets by index (normally of length <= 2)
	targets []int
	// this gives the index for the  best member in Pt
	bestMember int
	// heuristics for the group
	hu *PsoHeuristics
}

// Pso is the Particle swarm optimizer
type Pso struct {
	// random number generator
	rnd *rand.Rand
	// collection of particles
	Pt []Particle
	// mapped collection of groups of particles (with same heuristic settings)
	gr map[string]*Group
	//scratch pad for intermediate parameter calculations
	temp *big.Int
	//scratch pad for intermediate velocity calculation
	tempVel []float64
	// this gives the index of the particle with best try
	bestParticle int
	// cost function
	fun Fun
	// maximum number of elements in sets
	maxLen int
	// max integer; all parameters integer values less than this
	maxN *big.Int
	// number of particles
	n int
	// heuristics for the groups to derive from
	hu *PsoHeuristics
}

/*NominalL returns a nominal LfactorHeuristic value for  the LfactorHeuristic based on
* Number of particles and number of parameters. This was tuned for the
* continuous case this needs tuning for  the discrete  set case and is not yet
* used.
 */
func (pso *Pso) NominalL() (l float64) {
	const a1 = 0.91
	const a2 = 0.21
	const a3 = 0.51
	const a4 = 0.58

	return a1 * a3 / (math.Pow(float64(pso.n), a2) * math.Pow(float64(pso.maxLen), a4))
}

/*
NewPso sets up a PSO with a swarm of n particles. fun is the costing function
interface. Each Particle has a big integer Parameter. To initialize the
parameter set values are uniformly randomly chosen with integer values up to
Func.MaxLen()bits; ToConstraint() is used on these random choices with repeated
use  of random choice until each Particle has an initial Parameters that
satisfies the cost function  constraints on the Parameters. Default heuristics
are applied  to the "root" Group and all particles are added to this group.
The PSO uses the random generator seed sd.
*/
func NewPso(n int, fun Fun,
	sd int64) *Pso {
	var pso Pso
	pso.rnd = rand.New(rand.NewSource(sd))
	pso.n = n
	pso.maxLen = fun.MaxLen()
	pso.fun = fun
	pso.temp = big.NewInt(0)
	pso.tempVel = make([]float64, pso.maxLen)
	pso.gr = make(map[string]*Group, n)
	g := new(Group)
	pso.gr["root"] = g
	g.id = "root"
	pso.hu = pso.CreatePsoHeuristics()
	pso.SetGroupHeuristics(g, pso.hu)
	g.members = make([]int, n)

	g.targets = make([]int, 1, 2)

	pso.Pt = make([]Particle, n)
	maxN := big.NewInt(0)
	maxN.SetBit(maxN, pso.maxLen, 1)
	pso.maxN = maxN
	for i := range pso.Pt {
		p := &pso.Pt[i]
		p.group = g
		p.hint = big.NewInt(0)
		p.current = pso.fun.NewTry()
		p.bestTry = pso.fun.NewTry()
		//search for state that satisfies function constraints
		searching := true
		for searching {
			p.hint.Rand(pso.rnd, maxN)
			searching = !pso.fun.ToConstraint(p.current, p.hint)
			//fmt.Printf("part= %d param= %v  %s %s \n", i, p.current.Parameter(), p.current.Decode(), p.current.Cost())
		}
		pso.fun.Copy(p.bestTry, p.current)
		g.members[i] = i
		p.vel = make([]float64, pso.maxLen)

	}
	pso.UpdateGlobal()
	return &pso
}

//Part returns  ith particle
func (pso *Pso) Part(i int) *Particle {
	return &pso.Pt[i]
}

/*UpdateGroup does housekeeping for the group k
it calculates the best try and the particle in group that gives this.
Note that it looks for the best in the current iteration and disregards historic
best costs even if they were better.
*/
func (pso *Pso) UpdateGroup(g *Group) {
	if len(g.members) == 0 {
		g.bestMember = -1
		return
	}
	g.bestMember = g.members[0]
	for i := range g.members {
		id := g.members[i]
		result := pso.fun.Cmp(pso.Pt[g.bestMember].bestTry,
			pso.Pt[id].bestTry, futil.CostMode)
		if result > 0.0 {
			g.bestMember = id
		}
	}
	// if g.bestMember != 0 {
	// 	fmt.Printf("id=%d ", g.bestMember)
	// }
}

/*UpdateGlobal updates the global best by calling UpdateGroup()
and then finding the best group with its best particle.
 It searches for minimum best cost particle.
*/
func (pso *Pso) UpdateGlobal() {
	for _, g := range pso.gr {
		pso.UpdateGroup(g)
	}
	pso.bestParticle = 0
	for _, g := range pso.gr {
		if len(g.members) > 0 {
			//fmt.Printf("bestmember= %d", g.bestMember)
			compResult := pso.fun.Cmp(pso.Pt[pso.bestParticle].bestTry,
				pso.Pt[g.bestMember].bestTry, futil.CostMode)
			if compResult > 0.0 {
				pso.bestParticle = g.bestMember

			}

		}
	}
}

//BestParticle returns the current global best particle.
func (pso *Pso) BestParticle() int {
	return pso.bestParticle
}

// Nparticles returns the number of particles in use.
func (pso *Pso) Nparticles() int {
	return len(pso.Pt)
}

//CurrentTry returns the current try for the ith particle
func (pso *Pso) CurrentTry(i int) Try {
	return pso.Pt[i].current
}

//LocalBestTry returns Local-best try of the ith Particle
func (pso *Pso) LocalBestTry(i int) Try {
	return pso.Pt[i].bestTry
}

//Heuristics returns a copy of the master heuristics
func (pso *Pso) Heuristics() *PsoHeuristics { return pso.hu }

// CardinalSize returns the number of 1's in a binary representation of Int
// (assuming it is positive) so is the cardinal size of the corresponding set.
func CardinalSize(x *big.Int) (card int) {
	card = 0
	j := x.BitLen()
	for i := 0; i < j; i++ {
		if x.Bit(i) == 1 {
			card++
		}
	}
	return
}

/*
BlurTarget blurs the target set displacement based on its CardinalSize. Normally
x is the exclusive or of a particle Parameters with a target Local-best
Parameters. The blur is effected indirectly  by pseudo adding a probability

  pb =rand*(l*CardinalSize(x)+l0)/MaxLen()

to each component of the Velocity;l,l0 corresponds to the LfactorHeuristic, LoffsetHeuristic
heuristics;MaxLen() is the maximum number of set items ;rand is a random number
between 0 and 1.
*/
func (pso *Pso) BlurTarget(x *big.Int, id int, l, l0 float64) {
	h := l*float64(CardinalSize(x)) + l0
	p := &pso.Pt[id]
	// add blur via velocity increment
	prob := pso.rnd.Float64() * h / float64(pso.maxLen)
	for i := range p.vel {
		p.vel[i] = p.vel[i]*(1.0-prob) + prob
	}
}

/*
SetParams sets the parameters of the id th particle updating  the resulting
cost and  personal best case. It also revaluates the personal best cost in
case the function has changed.
*/
func (pso *Pso) SetParams(id int) {
	p := &pso.Pt[id]
	pso.fun.UpdateCost(p.bestTry)
	// update cost if the hint can be converted to a constraint satisfying
	// subset
	if pso.fun.ToConstraint(p.current, p.hint) {
		// if p.debug {
		// 	fmt.Printf("constraint update part= %d  %s %s \n", id, p.bestTry.Decode(), p.bestTry.Cost())
		// 	p.debug = false
		// }

		for i := range p.tries {
			pso.fun.UpdateCost(p.tries[i])
		}
		p.lookForBetterTry(pso)
		compResult := pso.fun.Cmp(p.bestTry, p.current, futil.CostMode)
		//fmt.Printf("compResult = %f \n", compResult)

		if compResult > pso.hu.Float(ThresholdHeuristic) {
			//p.putOntoTryList(pso, p.bestTry)
			pso.fun.Copy(p.bestTry, p.current)
			//fmt.Printf("part= %d  %s %s \n", id, p.bestTry.Decode(), p.bestTry.Cost())
			// if p.bestTry.Fbits() < 8.2 {
			// 	fmt.Printf("part= %d  %s %s \n", id, p.bestTry.Decode(), p.bestTry.Cost())
			// 	p.debug = true
			// pso.fun.UpdateCost(p.bestTry)
			// fmt.Printf("update part= %d  %s %s \n", id, p.bestTry.Decode(), p.bestTry.Cost())

			//}

		} else if compResult > -pso.hu.Float(ThresholdHeuristic) {
			p.putOntoTryList(pso, p.current)
		}
	}
}

func (p *Particle) putOntoTryList(pso *Pso, t Try) {
	try := pso.fun.NewTry()
	pso.fun.Copy(try, t)
	p.tries = append(p.tries, try)
	if len(p.tries) > pso.hu.Int(NTriesHeuristic) {
		p.removeWorstTry(pso)
	}
}

func (p *Particle) lookForBetterTry(pso *Pso) {
	// if len(p.tries)>0{
	// 	fmt.Printf("trys len = %d \n",len(p.tries))
	// }
	j := -1
	betterResult := 0.0
	for i := range p.tries {
		result := pso.fun.Cmp(p.bestTry, p.tries[i], futil.TriesMode)
		if result > betterResult {
			j = i
			betterResult = result
		}
	}
	if j >= 0 && betterResult > 1.0 {
		pso.fun.Copy(p.bestTry, p.tries[j])
		p.tries = append(p.tries[:j], p.tries[j+1:]...)
	}
}

func (p *Particle) removeWorstTry(pso *Pso) {
	j := -1
	worstResult := math.MaxFloat64
	for i := range p.tries {
		result := pso.fun.Cmp(p.bestTry, p.tries[i], futil.TriesMode)
		if result < worstResult {
			j = i
			worstResult = result
		}
	}
	if j >= 0 {
		p.tries = append(p.tries[:j], p.tries[j+1:]...)
	}
}

func (pso *Pso) setTempVel(z *big.Int, prob float64) {
	for i := range pso.tempVel {
		if z.Bit(i) > 0 {
			pso.tempVel[i] = prob
		} else {
			pso.tempVel[i] = 0.0
		}
	}
}
func (pso *Pso) addToTempVel(z *big.Int, prob float64) {
	for i := range pso.tempVel {
		if z.Bit(i) > 0 {
			pso.tempVel[i] = pso.tempVel[i]*(1.0-prob) + prob
		}
	}
}
func (p *Particle) addToVel(z *big.Int, prob float64) {
	for i := range p.vel {
		if z.Bit(i) > 0 {
			p.vel[i] = p.vel[i]*(1.0-prob) + prob
		}
	}
}

/*
PUpdate does a single step update assuming that groups have been setup with
the appropriate targets and heuristics.

for each Particle  the Parameters difference between itself Personal-best and
Targets is represented by taking the exclusive or of the corresponding
Parameters as subset. This is then used to pseudo add up contributions to target
blur prior to reducing the velocity  by a factor equal to the OmegaHeuristic heuristic.

Then the PhiHeuristic is used to generate randomly chosen parameters r
for each exclusive or between 0 and PhiHeuristic; each r is checked to be <= 1.0 and is
forced  to be so by replacing by 2-r if it isn't. These r are then pseudo added
to the  the velocity components that have an exclusive or bit of 1.

In this way the Velocity components are computed to encourage movement toward
Personal-best and Targets after encouraging more mutation for distant Targets
and Personal-best Parameters.

After this Setparams() and UpdateGlobal() are called to finish the update.
*/
func (pso *Pso) PUpdate() {
	for k := range pso.Pt {
		p := &pso.Pt[k]
		g := p.group
		PhiHeuristic := g.hu.Float(PhiHeuristic)
		l := g.hu.Float(LfactorHeuristic)
		l0 := g.hu.Float(LoffsetHeuristic)

		pso.temp.Xor(p.bestTry.Parameter(), p.current.Parameter())
		pso.BlurTarget(pso.temp, k, l, l0)
		rp := PhiHeuristic * pso.rnd.Float64()
		if rp > 1 {
			rp = 2 - rp
		}
		// add personal best velocity contribution
		pso.setTempVel(pso.temp, rp)
		for t := range g.targets {
			target := g.targets[t]
			pso.temp.Xor(pso.Pt[target].bestTry.Parameter(),
				p.current.Parameter())
			pso.BlurTarget(pso.temp, k, l, l0)
			rg := PhiHeuristic * pso.rnd.Float64()
			if rg > 1 {
				rg = 2 - rg
			}
			// add target best velocity contribution
			pso.addToTempVel(pso.temp, rg)
		}
		//reduce velocity and then combine contributions
		OmegaHeuristic := g.hu.Float(OmegaHeuristic)
		for jv := range p.vel {
			p.vel[jv] *= OmegaHeuristic
			p.vel[jv] = (1.0-pso.tempVel[jv])*p.vel[jv] + pso.tempVel[jv]
		}

		// update parameter
		p.hint.Set(p.current.Parameter())
		for jv := range p.vel {
			if pso.rnd.Float64() < p.vel[jv] {
				p.vel[jv] = 0.0
				if p.hint.Bit(jv) == 0 {
					p.hint.SetBit(p.hint, jv, 1)
				} else {
					p.hint.SetBit(p.hint, jv, 0)
				}
			}
		}
	}
	for i := range pso.Pt {
		pso.SetParams(i)
	}
	pso.UpdateGlobal()
}

//CreateGroup creates a group named 'name' with a slot for 'ntargets' targets.
//The heuristics for the group are shared from the 'root' group.
func (pso *Pso) CreateGroup(name string, ntargets int) *Group {
	g := new(Group)
	g0 := pso.Gr("root")
	g.hu = g0.hu
	g.targets = make([]int, ntargets)
	pso.gr[name] = g
	g.id = name
	return g
}

// Gr returns group given by name it returns nil if it does not exist
func (pso *Pso) Gr(name string) *Group {
	return pso.gr[name]
}

//Group returns group for particle id
func (pso *Pso) Group(id int) *Group {
	return pso.Pt[id].group
}

//MoveTo moves the particle 'pat' to the designated group
func (pso *Pso) MoveTo(g *Group, pat int) {
	g0 := pso.Pt[pat].group
	// find pat in group members and delete
	j := -1
	for i, m := range g0.members {
		if m == pat {
			j = i
			break
		}
	}
	l := len(g0.members)
	for ; j < l-1; j++ {
		g0.members[j] = g0.members[j+1]
	}
	g0.members = g0.members[:l]
	g.members = append(g.members, pat)
	pso.Pt[pat].group = g

}

/*
PsoHeuristics is the collection of group heuristics, where some of the
heuristics may not be used. The default values are in brackets.

Support for future heuristics

Future heuristic parameters may be added to this list which can be ignored by
earlier PSO variants by suitable choice of defaults. Note heuristics are often
shared between groups so it is important to know where this is done when
updating a group's  heuristics.
*/
type PsoHeuristics struct {
	// storage of float values
	floatValues []float64
	// storage of int values
	intValues []int
}

/*
CreatePsoHeuristics returns a pointer to a SPSO heuristics parameters with
default values. pso is used to pass SPSO instance parameters such as number of particles and the initial number of elements in the parameter set to help provide tuned heuristics.
*/
func (pso *Pso) CreatePsoHeuristics() *PsoHeuristics {
	hu := new(PsoHeuristics)
	hu.floatValues = make([]float64, numberOfFloatHeuristics)
	hu.intValues = make([]int, numberOfIntHeuristics)
	hu.ToDefault(pso)
	return hu
}

/*
ToDefault sets the heuristics to their default value. Note this may change
in future when better defaults are found. pso is used to pass SPSO parameters.
*/
func (hu *PsoHeuristics) ToDefault(pso *Pso) {
	hu.SetFloat(PhiHeuristic, 1)
	hu.SetFloat(OmegaHeuristic, 0.73)
	hu.SetFloat(LfactorHeuristic, 0.15)
	hu.SetFloat(LoffsetHeuristic, 2.0)
	hu.SetFloat(ThresholdHeuristic, 0.99)
	hu.SetInt(NTriesHeuristic, 250)

	hu.SetInt(TryGapHeuristic, 100)
}

const ( // floating  point  heuristics indexes
	//PhiHeuristic for target shooting probability range (1.0 )
	PhiHeuristic = iota
	//OmegaHeuristic for probability velocity factoring after target blur (0.73)
	OmegaHeuristic = iota
	//LfactorHeuristic for target blur factor (0.15)
	LfactorHeuristic = iota
	//LoffsetHeuristic for target blur offset (2.0)
	LoffsetHeuristic = iota
	//ThresholdHeuristic for acting on a comparison(0.99)
	ThresholdHeuristic      = iota
	numberOfFloatHeuristics = iota
)

const ( //integer heuristics indexes
	//NTriesHeuristic for  maximum number of tries  stored in a particle(250)
	NTriesHeuristic = iota
	//TryGapHeuristic for a minimum number of tries before doing something different(100)
	TryGapHeuristic       = iota
	numberOfIntHeuristics = iota
)

// Float returns the ith floating point heuristic
func (hu *PsoHeuristics) Float(i int) float64 {
	return hu.floatValues[i]
}

// SetFloat sets the ith floating point heuristic to x
func (hu *PsoHeuristics) SetFloat(i int, x float64) {
	hu.floatValues[i] = x
}

// Int returns the ith floating point heuristic
func (hu *PsoHeuristics) Int(i int) int {
	return hu.intValues[i]
}

// SetInt sets the ith integer heuristic to x
func (hu *PsoHeuristics) SetInt(i, x int) {
	hu.intValues[i] = x
}

/*
Heuristics returns the pointer to the group's heuristics
*/
func (g *Group) Heuristics() (hu *PsoHeuristics) {
	return g.hu
}

/*
SetGroupHeuristics sets the heuristics for the group g. As the Pso undergoes
development additional heuristic parameters can be introduced with compatible
defaults without destroying backwards compatibility with previous versions that
do not have the extra heuristics. Not that only a link to the heuristics is
passed so the heuristics can be changed outside the group with many groups
referring to the  same heuristics.

All group heuristics are derived, normally by just a pointer, from a master
heuristics stored in pso, so often all heuristics  can be changed via the master
heuristic; how this is done should be through calling SetHeuristic() for the
derived SPSO.
*/
func (pso *Pso) SetGroupHeuristics(g *Group, h *PsoHeuristics) {
	g.hu = h
}

// SetGroupTarget sets the first few Targets of group 'grp'
// to the particle list targetList.
func (pso *Pso) SetGroupTarget(grp *Group, targetList ...int) {

	copy(grp.targets, targetList)
}

// GroupBest returns the best particle id in the group grp.
// It returns -1 if there is no member.
func (pso *Pso) GroupBest(grp *Group) (best int) {
	return grp.bestMember
}

// GPso is a PSO that targets the global best element and has only one Group
// corresponding to all the particles
type GPso struct {
	*Pso
}

// NewGPso creates a GPso
func NewGPso(p *Pso) *GPso {
	gp := &GPso{p}
	return gp
}

//SetHeuristics sets the heuristics for the particle swarm.
func (p *GPso) SetHeuristics(hu *PsoHeuristics) { p.hu = hu }

// Update sets the target to the current global best  Particle before updating
// the particle swarm
func (p *GPso) Update() {
	g := p.Group(0)
	p.SetGroupTarget(g, p.GroupBest(g))
	p.PUpdate()
}

/*CLPso is a comprehensive learning PSO that targets other personal best
rather than just the global best. Each particle has its own group with group name
set to the particles id and has one target. All particles share Heuristics.
*/
type CLPso struct {
	*Pso
	clPt []CLpart
	//target refreshing gap
	TryGap int
}

//CLpart includes the additional particle state used by CLPso to manage particle
// targets.
type CLpart struct {
	// probability of learning from other particles
	pc float64
	// last best try while pursuing the current target
	lastBest Try
	// failure to improve count while pursuing current target
	gapCount int
}

/*
Pc0 calculates the assigned probability of learning from other particles. i  is the
Particle id and n is the number of particles. The formula is an empirical one
used in the continuous case. This is likely to be replaced by another
empirical formula called Pc1 and Pc0 is just a stand in for Pc1.
*/
func Pc0(i int, n float64) float64 {
	e0 := (math.Exp(10.0) - 1.0)
	x := (math.Exp(10*float64(i)/(n-1.0)) - 1.0) / e0
	return 0.05 + 0.45*x
}

/*
NewCLPso creates a CLPso and sets the heuristics if required. In this case
TryGapHeuristic is the number of iterations with no improvement needed to
trigger the probabilistic search for an alternative target. The ith particle is
assigned the probability Pc0(i,n)  to look for an alternative target when
triggered, where n is the number of particles.
*/
func NewCLPso(p *Pso) *CLPso {
	clpt := make([]CLpart, p.Nparticles())
	pso := &CLPso{p, clpt, 0}
	n := float64(pso.Nparticles())
	pso.TryGap = pso.hu.Int(TryGapHeuristic)
	for i := range pso.clPt {
		c := &pso.clPt[i]
		c.pc = Pc0(i, n)
		c.lastBest = pso.fun.NewTry()
		c.gapCount = -1 // play safe
		gp := pso.CreateGroup(string(i), 1)
		pso.MoveTo(gp, i)
	}
	return pso
}

//SetHeuristics sets the heuristics for the particle swarm.
func (p *CLPso) SetHeuristics(hu *PsoHeuristics) { p.hu = hu }

/*
Update does the iteration update. For each Particle a check is made to see if
the updates have not given an improvement in cost of the Parameters compared to
the last improvement for at least TryGap iterations. If so it looks for an
alternative target with probability pc=Pc0(). If it looks for an alternative
target it randomly selects two Particles and chooses the one  that gives the
least of the Local-best Cost. Note the choice is from all particles and can
include itself. After this it does the usual PUpdate().
*/
func (p *CLPso) Update() {
	for i := range p.clPt {
		c := &p.clPt[i]
		p.fun.UpdateCost(c.lastBest)
		g := p.Group(i)
		try := p.Pt[i].current
		if c.gapCount > p.TryGap {
			c.gapCount = 0
			p.fun.Copy(c.lastBest, try)
			if p.rnd.Float64() < c.pc {
				i1 := p.rnd.Intn(p.Nparticles())
				i2 := p.rnd.Intn(p.Nparticles())
				compResult := p.fun.Cmp(p.Pt[i1].bestTry, p.Pt[i2].bestTry, futil.CostMode)
				if compResult > 0.0 {
					p.SetGroupTarget(g, i2)
				} else {
					p.SetGroupTarget(g, i1)
				}
			} else {
				p.SetGroupTarget(g, i)
			}
		} else {
			if c.gapCount < 0 { // needs initializing
				p.fun.Copy(c.lastBest, try)
				c.gapCount = p.TryGap + 1
			} else if p.fun.Cmp(c.lastBest, try, futil.CostMode) > 0.0 {
				p.fun.Copy(c.lastBest, try)

				c.gapCount = 0
			} else {
				c.gapCount++
			}
		}
	}
	p.PUpdate()
}

/*PrintDebug outputs debugging diagnostics depending on the command id.
It has the following values and output:
  Command | Output
	===============================
	group   | group data for each Particle
	group0  | "root" group data's best member and cost
	Pt      | particle local-best cost and parameter and
				  |  current parameter for each particle
	vel     | velocity for each particle

*/
func (pso *Pso) PrintDebug(w io.Writer, id string) {
	switch id {
	case "group":
		fmt.Fprintf(w, "group data: \n")
		for i := range pso.Pt {
			fmt.Fprintf(w, " %d group = %v \n", i, *pso.Pt[i].group)

		}
	case "group0":
		g := pso.gr["root"]
		fmt.Fprintf(w, "best member = %d,cost = %v \n",
			g.bestMember, pso.Pt[g.bestMember].bestTry.Cost())
	case "Pt":
		for i := range pso.Pt {
			p := &pso.Pt[i]
			try := p.bestTry
			fmt.Fprintf(w, "%d bestcost= %v\n", i, try.Cost())
			fmt.Fprintf(w, "bestparam = %s\n", try.Parameter().Text(2))
			fmt.Fprintf(w, "param = %s\n", p.current.Parameter().Text(2))
		}
	case "vel":

		for i := range pso.Pt {
			p := &pso.Pt[i]
			cnt := 0
			fmt.Fprintf(w, "vel %d\n", i)
			for j := range p.vel {
				fmt.Fprintf(w, "  %f", p.vel[j])
				cnt++
				if cnt == 5 {
					fmt.Fprintf(w, "\n")
					cnt = 0
				}
			}
		}
		fmt.Fprintf(w, "\n")
	}

}
