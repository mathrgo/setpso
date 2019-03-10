package setpso

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
)

/*
Fun is the interface to the function that calculates the cost and the number
of bits used in the binary pattern. The binary pattern is stored as a big
integer. The integer can be regarded as a sub set of integers from 0 to
MaxLen()-1 where i in this range is a member of the sub set if the ith bit is
1. Cost() gives the cost associated by the function to a binary pattern
(stored as a big integer). MaxLen() returns  the maximum number of bits in
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
combinatorial optimisation so the cost function must provide an easily
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
	// evaluated cost where a lower cost is better
	Cost(x *big.Int) *big.Int
	// maximum number of bits in the parameter big integer which is the
	// maximum number of elements in the subset
	MaxLen() (maxlen int)
	// string giving a description of the cost function
	About() (s string)
	// this attempts to  give a constraint satisfying hint that matches the hint;
	// pre is the previous constraint satisfying version to hint, which
	// should not be changed
	ToConstraint(pre, hint *big.Int) bool
	// this requests the function to give a meaningful interpretation of
	// z as a Parameters subset for the function assuming z satisfies constraints
	Decode(z *big.Int) (s string)
	// this hints to the function to remove/replace the ith item
	Delete(i int) bool
}

/*
PsoInterface is the interface used with the PSO  and its variants. Normally
Pso supplies this apart from the Udate() function. The interface is used by
the monitoring and running program  and psokit uses this interface to provide
useful plotting functions.
*/
type PsoInterface interface {
	// single state update to all particles in the PSO
	Update()
	// best global Parameters so far
	GlobalParams() *big.Int
	// the cost of the global best
	GlobalCost() *big.Int
	// number of particles
	Nparticles() int
	// current parameters of the ith Particle
	Params(i int) *big.Int
	// Local-best cost of the ith Particle
	LocalBestCost(i int) *big.Int
	// Local-best Prameters of the ith Particle
	LocalBestParameters(i int) *big.Int
	// interface for requesting debug output based on cmd
	PrintDebug(w io.Writer, cmd string)
	//Heuristics returns a copy of the master heuristics
	Heuristics() PsoHeuristics
	//SetHeuristics sets the heuristics for the particle swarm.
	SetHeuristics(hu PsoHeuristics)
}

// Particle is the state of a member of the PSO
type Particle struct {
	// current parameters normalised to have similar effect and search area
	params *big.Int
	// this is used as a raw update before mapping to a constraint satisfying
	// state
	hint *big.Int
	// current computed cost
	cost *big.Int
	// pointer to group the particle belongs to
	group *Group
	// best parameter values so far
	bestParams *big.Int
	//  best personal cost so far corresponding to this parameter
	best *big.Int
	// current  flipping probability requests for each bit component
	vel []float64
}

//Group is a collection of particles with same heuristic settings
type Group struct {
	//group's id
	id string
	// array of member particles
	members []int
	// list of update targets by index (normally of length <= 2)
	targets []int
	// this gives the cost of the best member
	bestCost *big.Int
	// this gives the index for the  best member in Pt
	bestMember int
	// heuristics for the group
	hu *PsoHeuristics
}

// Pso is the Particle swarm optimiser
type Pso struct {
	// random number generator
	rnd *rand.Rand
	// collection of particles
	Pt []Particle
	// mapped collection of groups of particles (with same heuristic settings)
	gr map[string]*Group
	// group id with best cost
	bestGroup *Group
	//scratch pad for intermediate parameter calculations
	temp *big.Int
	//scratch pad for intermediate velocity calculation
	tempVel []float64
	// this gives the cost of the global best
	bestCost *big.Int
	// this gives the parameters for the global best
	bestParams *big.Int
	// cost function
	fun Fun
	// maximum number of elements in sets
	maxLen int
	// max integer?
	maxN *big.Int
	// number of particles
	n int
	// heuristics for the groups to derive from
	hu PsoHeuristics
}

/*NominalL returns a nominal Lfactor value for  the heuristic HLfactor based on
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

/*NewPso sets up a PSO with a swarm of n particles. fun is the costing function
* interface. Each Particle has a big integer Parameter. To initialise the
* parameter set values are uniformly randomly chosen with integer values upto
* Func.MaxLen(); ToConstraint() is used on these random choices with repeated
* use  of random choice until each Particle has an initial Parameters that
* satisfies the cost function  constraints on the Parameters. Default heuristics
* are applied  to the "root" Group and all particles are added to this group.
* The PSO uses the random generator seed sd.
 */
func NewPso(n int, fun Fun,
	sd int64) *Pso {
	var pso Pso
	pso.rnd = rand.New(rand.NewSource(sd))
	pso.n = n
	pso.hu.ToDefault()
	pso.maxLen = fun.MaxLen()
	pso.fun = fun
	pso.temp = big.NewInt(0)
	pso.tempVel = make([]float64, pso.maxLen)
	pso.gr = make(map[string]*Group, n)
	g := new(Group)
	pso.gr["root"] = g
	g.id = "root"

	pso.SetGroupHeuristics(g, &pso.hu)
	g.members = make([]int, n)

	g.targets = make([]int, 1, 2)

	g.bestCost = big.NewInt(0)

	pso.bestCost = big.NewInt(0)
	pso.bestParams = big.NewInt(0)
	pso.Pt = make([]Particle, n)
	maxN := big.NewInt(0)
	maxN.SetBit(maxN, pso.maxLen, 1)
	pso.maxN = maxN
	for i := range pso.Pt {
		p := &pso.Pt[i]
		p.group = g
		p.params = big.NewInt(0)
		p.hint = big.NewInt(0)
		//serch for state that satisfies function constraints
		serching := true
		for serching {
			p.hint.Rand(pso.rnd, maxN)
			serching = !pso.fun.ToConstraint(p.params, p.hint)
		}
		p.params.Set(p.hint)
		p.cost = big.NewInt(0)
		g.members[i] = i
		p.cost.Set((pso.fun).Cost(p.params))
		p.best = big.NewInt(0)
		p.best.Set(p.cost)
		p.bestParams = big.NewInt(0)
		p.bestParams.Set(p.params)
		p.vel = make([]float64, pso.maxLen)

		//fmt.Printf(" %d cost= %v\n", i, p.best)
	}
	pso.UpdateGlobal()
	return &pso
}

/*UpdateGroup does housekeeping for the group k
it calculates the best cost and the particle in group that gives this.
Note that it looks for the best in the current iteration and disregards historic
best costs even if they were better.
*/
func (pso *Pso) UpdateGroup(g *Group) {
	g.bestCost.Set(pso.maxN)
	g.bestMember = -1
	for i := range g.members {
		id := g.members[i]
		p := &pso.Pt[id]
		if p.best.Cmp(g.bestCost) < 0 {
			g.bestCost.Set(p.best)
			g.bestMember = id
		}
	}
}

/*UpdateGlobal updates the global best by calling UpdateGroup()
and then finding the best group with its best particle.
 It searches for minimum best cost particle.
*/
func (pso *Pso) UpdateGlobal() {
	pso.bestCost.Set(pso.maxN)
	for _, g := range pso.gr {
		pso.UpdateGroup(g)
	}
	pso.bestGroup = nil
	// assume pso.gr is not empty
	for _, g := range pso.gr {
		if g.bestCost.Cmp(pso.bestCost) < 0 {
			pso.bestCost.Set(g.bestCost)
			pso.bestGroup = g
		}
	}
	//copy across the best case
	pso.bestParams.Set(pso.Pt[pso.bestGroup.bestMember].bestParams)
}

// GlobalCost returns the current global best cost.
func (pso *Pso) GlobalCost() *big.Int {
	return pso.bestCost
}

// Nparticles returns the number of particles in use.
func (pso *Pso) Nparticles() int {
	return len(pso.Pt)
}

//Params returns the current parameters for the ith particle.
func (pso *Pso) Params(i int) *big.Int {
	return pso.Pt[i].params
}

//LocalBestCost returns the current local best cost  for the ith particle.
func (pso *Pso) LocalBestCost(i int) *big.Int {
	return pso.Pt[i].best
}

//LocalBestParameters returns the the Parameters of the Local-best for the
// ith Particle.
func (pso *Pso) LocalBestParameters(i int) *big.Int {
	return pso.Pt[i].bestParams
}

// GlobalParams returns the parameters for the global best cost.
func (pso *Pso) GlobalParams() *big.Int {
	return pso.bestParams
}

//Heuristics returns a copy of the master heuristics
func (pso *Pso) Heuristics() PsoHeuristics { return pso.hu }

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
RandomSubset converts z to a subset of z by using a probability p for
selecting each member of z as a set. the updated z is returned using the pso
random number generator.
*/
func RandomSubset(z *big.Int, p float64, rnd *rand.Rand) *big.Int {
	n := z.BitLen()
	for i := 0; i < n; i++ {
		if z.Bit(i) == 1 {
			if rnd.Float64() > p {
				z.SetBit(z, i, 0)
			}
		}
	}
	return z
}

//ToggleRandomSet takes the set z and randomly toggles in and out
// m members out of n members and returns z.
func ToggleRandomSet(z *big.Int, m int, n int, rnd *rand.Rand) *big.Int {
	for i := 0; i < m; i++ {
		j := rnd.Intn(n)
		b := z.Bit(j)
		if b == 0 {
			b = 1
		} else {
			b = 0
		}
		z.SetBit(z, j, b)
	}
	return z
}

/*
BlurTarget blurs the target set displacement based on its CardinalSize. Normally
x is the exclusive or of a particle Parameters with a target Local-best
Parameters. The blur is effected indirectly  by pseudo adding a probability

  pb =rand*(l*CardinalSize(x)+l0)/MaxLen()

to each component of the Velocity;l,l0 corresponds to the Lfactor, Loffset
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
cost and  personal best case. It also reevaluates the personal best cost in
case the function has changed.
*/
func (pso *Pso) SetParams(id int) {
	p := &pso.Pt[id]
	// update cost if the hint can be converted to a constraint satisfying
	// subset
	if pso.fun.ToConstraint(p.params, p.hint) {
		p.params.Set(p.hint)
		p.best.Set(pso.fun.Cost(p.bestParams))
		p.cost.Set(pso.fun.Cost(p.params))
		if p.cost.Cmp(p.best) < 0 {
			p.best.Set(p.cost)
			p.bestParams.Set(p.params)
		}
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
blur prior to reducing the velocity  by a factor equal to the Omega heuristic.

Then the Phi heuristic is used to generate randomly chosen parameters r
for each exclusive or between 0 and Phi; each r is checked to be <= 1.0 and is
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
		Phi := g.hu.Phi
		l := g.hu.Lfactor
		l0 := g.hu.Loffset

		pso.temp.Xor(p.bestParams, p.params)
		pso.BlurTarget(pso.temp, k, l, l0)
		rp := Phi * pso.rnd.Float64()
		if rp > 1 {
			rp = 2 - rp
		}
		// add personal best velocity contribution
		pso.setTempVel(pso.temp, rp)
		for t := range g.targets {
			targ := g.targets[t]
			pso.temp.Xor(pso.Pt[targ].bestParams, p.params)
			pso.BlurTarget(pso.temp, k, l, l0)
			rg := Phi * pso.rnd.Float64()
			if rg > 1 {
				rg = 2 - rg
			}
			// add target best velocity contribution
			pso.addToTempVel(pso.temp, rg)
		}
		//reduce velocity and then combine contributions
		Omega := g.hu.Omega
		for jv := range p.vel {
			p.vel[jv] *= Omega
			p.vel[jv] = (1.0-pso.tempVel[jv])*p.vel[jv] + pso.tempVel[jv]
		}

		// update parameter
		p.hint.Set(p.params)
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
	g.bestCost = big.NewInt(0) // this will be over written before use
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
	//for target shooting probability range (1.0 )
	Phi float64
	//for probability velocity factoring after target blur (0.73)
	Omega float64
	// for target blur factor (0.15)
	Lfactor float64
	// for taret blur offset (2.0)
	Loffset float64
	// for a minimum number of tries before doing something different(100)
	TryGap int
}

/*
ToDefault sets the heuristics to their default value. Note this may change
in future when better defaults are found.
*/
func (hu *PsoHeuristics) ToDefault() {
	hu.Phi = 1
	hu.Omega = 0.73
	hu.Lfactor = 0.15
	hu.Loffset = 2.0
	hu.TryGap = 100
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
do not have the extra heuristics. Not that only a link to the huristics is
passed so the heuristics can be changed outside the group with many groups
referring to the  same heuristics.

All group heuristics are derived, normaly by just a pointer, from a master
heuristics stored in pso, so often all heuristics  can be changed via the master
heuristic; how this is done should be through calling SetHeuristic() for the
derived SPSO.
*/
func (pso *Pso) SetGroupHeuristics(g *Group, h *PsoHeuristics) {
	g.hu = h
}

// SetGroupTarget sets the first fiew Targets of group 'grp'
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
func (p *GPso) SetHeuristics(hu PsoHeuristics) { p.hu = hu }

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
	// last best while pursuing the current target
	lastBest *big.Int
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
TryGap heuristic is the number of iterations with no improvement needed to
trigger the probabilistic search for an alternative target. The ith particle is
assigned the probability Pc0(i,n)  to look for an alternative target when
triggered, where n is the number of particles.
*/
func NewCLPso(p *Pso) *CLPso {
	clpt := make([]CLpart, p.Nparticles())
	pso := &CLPso{p, clpt, 0}
	n := float64(pso.Nparticles())
	pso.TryGap = pso.hu.TryGap
	for i := range pso.clPt {
		c := &pso.clPt[i]
		c.pc = Pc0(i, n)
		c.lastBest = big.NewInt(0)
		c.gapCount = -1 // play safe
		gp := pso.CreateGroup(string(i), 1)
		pso.MoveTo(gp, i)
	}
	return pso
}

//SetHeuristics sets the heuristics for the particle swarm.
func (p *CLPso) SetHeuristics(hu PsoHeuristics) { p.hu = hu }

/*
Update does the iteration update. For each Particle a check is made to see if
the updates have not given an improvement in cost of the Parametrs compared to
the last improvement for at least TryGap iterations. If so it looks for an
alternative target with probability pc=Pc0(). If it looks for an alternative
target it randomly selects two Particles and chooses the one  that gives the
least of the Local-best Cost. Note the choice is from all particles and can
include itself. After this it does the usual PUpdate().
*/
func (pso *CLPso) Update() {
	for i := range pso.clPt {
		c := &pso.clPt[i]
		g := pso.Group(i)
		if c.gapCount > pso.TryGap {
			c.gapCount = 0
			if pso.rnd.Float64() < c.pc {
				i1 := pso.rnd.Intn(pso.Nparticles())
				i2 := pso.rnd.Intn(pso.Nparticles())
				if (pso.LocalBestCost(i1)).Cmp(pso.LocalBestCost(i2)) < 0 {
					pso.SetGroupTarget(g, i1)
				} else {
					pso.SetGroupTarget(g, i2)
				}
			} else {
				pso.SetGroupTarget(g, i)
			}
		} else {
			cost := pso.Pt[i].cost
			if c.gapCount < 0 {
				c.lastBest.Set(cost)
				c.gapCount++
			} else if c.lastBest.Cmp(cost) > 0 {
				c.lastBest.Set(cost)
				c.gapCount = 0
			} else {
				c.gapCount++
			}
		}
	}
	pso.PUpdate()
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
		fmt.Fprintf(w, "best member = %d,cost = %v \n", g.bestMember, g.bestCost)
	case "Pt":
		for i := range pso.Pt {
			p := &pso.Pt[i]
			fmt.Fprintf(w, "%d bcost= %v\n", i, p.best)
			fmt.Fprintf(w, "bparam = %s\n", p.bestParams.Text(2))
			fmt.Fprintf(w, "xparam = %s\n", p.params.Text(2))
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
