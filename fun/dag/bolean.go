package dag

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/mathrgo/setpso/fun/futil"
)

/*
FunBool is the cost-function for evaluating boolean DAG encoded as a (positive) big integer.
*/
type FunBool struct {
	// temporary store of dag decoder
	*Dag
	// temporary store of dag values
	nodeValues []uint
	//temporary store of input
	input *big.Int
	//temporary store of output
	output *big.Int
	// temporary store of cost value
	cost *big.Int
	// store of cost value for optimiser
	costValue *futil.IntCostValue

	// temporary store of difference between output and required output
	// cost
	difCost *big.Int
	// node operator
	opt OptBool
	// sample generator
	sampler SamplerBool

	//factor weight to include INode usage cost
	sizeCostFactor *big.Int
	// random sample size for evaluating output missmatch cost
	sampleSize int
	// random number generator
	rnd *rand.Rand
}

/*
OptBool is  interface for interior node operation using boolean operations . l and r are the left and right
inputs to the node, opt is a, positive integer,  operations code
*/
type OptBool interface {
	Opt
	// function to generate a node output l,r are the left and right inputs and opt
	// is the encoded operation to be carried out on l,r to return the result.
	Opt(l, r uint, opt *big.Int) uint
}

/*
Opt4Bool encodes the Boolean operation as a 2-dimesion table of bits. the
rows are indexed by the left binary in put l and the column by the right binary
input r; the value of the table entry gives the corresponding binary output o.
each op is thus represented by a 4 bit integer opt where o=opt[r+2*l] treated
as an array of bits this gives 16 possible operations.
*/
type Opt4Bool struct {
	// symbols for representing operations in a human readable form
	symbol [16]string
	// cost  of using node table
	NodeCost [16]int
}

/*
NewOpt4Bool creates a Opt4Bool and inserts default node cost values
that discourages input negation, input selection, and ignore input nodes.
It promotes or("|"),and("&"), exclusive or ("+") operation nodes .
*/
func NewOpt4Bool() *Opt4Bool {
	var o Opt4Bool
	o.symbol = [16]string{
		" T!", " |!", "!& ", " <!", "!|!", " >!", " + ", " &!",
		" & ", " +!", " > ", "!| ", " < ", "!&!", " | ", " T "}
	// set up default node use cost
	o.NodeCost = [16]int{
		8, 2, 4, 2, 4, 2, 1, 2,
		1, 2, 4, 4, 4, 4, 1, 8,
	}
	return &o
}

// Cost gives the cost of using a node
func (o *Opt4Bool) Cost(opt *big.Int) int {
	return o.NodeCost[int(opt.Int64())]
}

// About gives description of operation type
func (o *Opt4Bool) About() string {
	s := " general binary boolean operation"
	return s
}

// Decode gives a human readable discriptin of opt.
func (o *Opt4Bool) Decode(opt *big.Int) string {
	return o.symbol[int(opt.Int64())]
}

// BitSize returns Number of bits needed to store opt encoding.
func (o *Opt4Bool) BitSize() int {
	return 4
}

// Opt is function to generate a node output. l,r are the left and right inputs
// and opt is the encoded operation to be carried out on l,r to return the
// result.
func (o *Opt4Bool) Opt(l, r uint, opt *big.Int) uint {
	return opt.Bit(int(2*l + r))
}

/*
SamplerBool is an interface for sampling inputs and required output results. Where x is the input array;
y is the corresponding required output array of bits given by the sampler; both arrays stored as big integers; rnd is the random number generator
used to create samples.
*/
type SamplerBool interface {
	Sample(x, y *big.Int, rnd *rand.Rand)
	InputSize() int
	OutputSize() int
	About() string
}

// NewFunBool returns a new *FunBool ready to be used.
func NewFunBool(nnode, nbitslookback int, opt OptBool, sizeCostFactor int64,
	sampler SamplerBool, sampleSize int, rnd *rand.Rand) *FunBool {
	nvar := sampler.InputSize()
	nout := sampler.OutputSize()

	f := FunBool{NewDag(nvar, nnode, nout, nbitslookback, opt),
		make([]uint, nnode),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		futil.NewIntCostValue(),
		big.NewInt(0),
		opt,
		sampler,
		big.NewInt(sizeCostFactor),
		sampleSize,
		rnd}

	f.sizeCostFactor = big.NewInt(sizeCostFactor)
	f.sampleSize = sampleSize
	return &f
}

// SetSizeCostFactor sets factor weight to include INode usage cost
func (f *FunBool) SetSizeCostFactor(sizeCostFactor int64) {
	f.sizeCostFactor.SetInt64(sizeCostFactor)
}

//SizeCostFactor returns factor weight to include INode usage cost
func (f *FunBool) SizeCostFactor() *big.Int {
	return f.sizeCostFactor
}

/*
Cost evaluates cost, where a lower cost is better. In this case
where

	cost = node usage cost * sizeCostFactor
	       + number of output component match errors

using random samples.
*/
func (f *FunBool) Cost(x *big.Int) futil.CostValue {
	f.Idecode(x) // assume satisfies constraints
	f.difCost.SetInt64(0)
	f.cost.SetInt64(int64(f.structureCost))

	for j := 0; j < f.sampleSize; j++ {
		f.sampler.Sample(f.input, f.output, f.rnd)
		// evaluate dag nodes using f.input from sample
		for i := 0; i < f.usedNodes; i++ {
			nd := &f.INodes[i]
			var l uint
			switch nd.ltype {
			case ITypeVar:
				l = f.input.Bit(nd.lindex)
			case ITypeNode:
				l = f.nodeValues[nd.lindex]
			case ITypeConst:
				l = 1
			}
			var r uint
			switch nd.rtype {
			case ITypeVar:
				r = f.input.Bit(nd.rindex)
			case ITypeNode:
				r = f.nodeValues[nd.rindex]
			case ITypeConst:
				r = 1
			}
			f.nodeValues[i] = f.opt.Opt(l, r, nd.opt)
		}
		// calculate missmatch with output
		c := int64(0)
		var cb big.Int
		for i := 0; i < f.sampler.OutputSize(); i++ {
			if f.nodeValues[f.outNodes[i]] != f.output.Bit(i) {
				c++
			}
		}
		cb.SetInt64(c)
		f.difCost.Add(f.difCost, &cb)
	}
	f.difCost.Mul(f.difCost, f.sizeCostFactor)
	f.cost.Add(f.cost, f.difCost)
	f.costValue.Set(f.cost)

	return f.costValue
}

// About string gives a description of the cost function
func (f *FunBool) About() string {
	s := fmt.Sprintf("Dag using operation: %s\n", f.opt.About())
	s += fmt.Sprintf("Sampler: %s\n ", f.sampler.About())
	s += fmt.Sprintf("sample size %d\n", f.sampleSize)
	s += fmt.Sprintf("Dag node cost factor %d", f.sizeCostFactor)
	return s
}

//Decode gives human radable description of encoding
func (f *FunBool) Decode(z *big.Int) string {
	s := f.DecodeDag(z)
	return s
}

//ToConstraint attempts to  give a constraint satisfying hint that matches the hint;
// pre is the previous constraint satisfying version to hint, which
// should not be changed. It returns True if it succeeds.
func (f *FunBool) ToConstraint(pre, hint *big.Int) bool {
	ok := f.Idecode(hint)
	return ok
}

// Delete hints to the function to remove/replace the ith item
// it returns true if the function takes the hint
func (f *FunBool) Delete(i int) bool {
	return false
}

//NewCostValue creates zero cost value
func (f *FunBool) NewCostValue() futil.CostValue {
	return futil.NewIntCostValue()
}
