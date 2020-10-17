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
	Opt(l, r uint, opt uint) uint
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
func (o *Opt4Bool) Cost(opt uint) int {
	return o.NodeCost[opt]
}

// About gives description of operation type
func (o *Opt4Bool) About() string {
	s := " general binary boolean operation"
	return s
}

// Decode gives a human readable discriptin of opt.
func (o *Opt4Bool) Decode(opt uint) string {
	return o.symbol[opt]
}

// BitSize returns Number of bits needed to store opt encoding.
func (o *Opt4Bool) BitSize() int {
	return 4
}

// Opt is function to generate a node output. l,r are the left and right inputs
// and opt is the encoded operation to be carried out on l,r to return the
// result.
func (o *Opt4Bool) Opt(l, r uint, opt uint) uint {
	return uint(1&(opt>>(2*l + r)))
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
//BoolTry gives the try structure to use
type BoolTry = futil.IntTry
//BoolFunStub gives interface to setpso
type BoolFunStub = futil.IntFunStub
// NewFunBool returns a new *FunBool ready to be used.
func NewFunBool(nnode, nbitslookback int, opt OptBool, sizeCostFactor int64,
	sampler SamplerBool, sampleSize int, rnd *rand.Rand) *BoolFunStub {
	nvar := sampler.InputSize()
	nout := sampler.OutputSize()
	var f FunBool
	f.Dag = NewDag(nvar, nnode, nout, nbitslookback, opt) 
	// temporary store of dag values
	f.nodeValues=make([]uint, nnode)
	//temporary store of input
	f.input=big.NewInt(0)
	//temporary store of output
	f.output= big.NewInt(0)
	// temporary store of cost value
	f.cost= big.NewInt(0)
	

	// temporary store of difference between output and required output
	// cost
	f.difCost = big.NewInt(0)
	// node operator
	f.opt=opt
	// sample generator
	f.sampler=sampler

	//factor weight to include INode usage cost
	f.sizeCostFactor= big.NewInt(0)
	// random sample size for evaluating output missmatch cost
	f.sampleSize =sampleSize
	// random number generator
	f.rnd =rnd

	f.sizeCostFactor = big.NewInt(sizeCostFactor)
	f.sampleSize = sampleSize
	return futil.NewIntFunStub(&f)
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
func (f *FunBool) Cost(data TryData, cost *big.Int){
	d:=data.(*DTryData)
	f.difCost.SetInt64(0)
	cost.SetInt64(int64(d.structureCost))

	for j := 0; j < f.sampleSize; j++ {
		f.sampler.Sample(f.input, f.output, f.rnd)
		// evaluate dag nodes using f.input from sample
		for i := 0; i < d.usedNodes; i++ {
			nd := &d.INodes[i]
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
			
			if f.nodeValues[d.outNodes[i]] != f.output.Bit(i) {
				c++
			}
			//fmt.Printf("out%d= %d,%d,%d\n",j,f.nodeValues[d.outNodes[i]],d.outNodes[i],f.output.Bit(i))
		}
		cb.SetInt64(c)
		f.difCost.Add(f.difCost, &cb)
	}
	f.difCost.Mul(f.difCost, f.sizeCostFactor)
	cost.Add(cost, f.difCost)
}
//DefaultParam gives a default that satisfies constraints
func (f *FunBool) DefaultParam() *big.Int {
	return big.NewInt(0)
}


// About string gives a description of the cost function
func (f *FunBool) About() string {
	s := fmt.Sprintf("Dag using operation: %s\n", f.opt.About())
	s += fmt.Sprintf("Sampler: %s\n ", f.sampler.About())
	s += fmt.Sprintf("sample size %d\n", f.sampleSize)
	s += fmt.Sprintf("Dag node cost factor %d", f.sizeCostFactor)
	return s
}





// Delete hints to the function to remove/replace the ith item
// it returns true if the function takes the hint
func (f *FunBool) Delete(i int) bool {
	return false
}
