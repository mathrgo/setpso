/*
Package dag  provides the base for Directed Asynchronous Graphs functions with
two inputs per node. It also includes examples using this base.
*/
package dag

import (
	"fmt"
	"math/big"
	"math/rand"
)

/*
Dag is the method for encoding/decoding special type of Directed Asynchronous
Graph into a binary string with an array of input nodes  and an array of ordered
interior nodes.

Node Linking

For encoding/decoding conceptually The inputs are placed before the interior
nodes to make one contiguous  array and each interior node has a pair of offset
values of positive integers each obtained by adding  1 to a binary string of
length nBinOffset. The offset values are used to link each interior  node to two
inputs. if an offset reaches further than available slots the offset is taken to
be pointing to an element with value 1.

Interior Node Operation

Each interior node is allocated OptSize() bits to encode how the two inputs to the node
are operated on to give an output value. At this level how this is done is
opaque: an interface called Opt does this instead using the method Opt().
*/
type Dag struct {
	// number of inputs
	nVar int
	//current number of interior nodes
	nNode int
	// number of outputs
	nOut int
	// maximum number of bits in the look back offset
	nBitsLookback int
	// out put node index array
	outNodes []int
	// number of bits needed to store a node's operations code.
	optSize int
	// interior node array
	INodes []Node
	// number of used nodes
	usedNodes int
	// interface to node operation
	Opt Opt
	// sum of used node costs
	structureCost int
}

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
	// temporary store of cost
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
Opt is bearbones interface used by Dag to decode operations into human readable form for the nodes and
reserve spacing for the nodes operation encoding.
*/
type Opt interface {
	// gives a description of the operation type
	About() string
	//this gives a human readable  version of the encoded operation opt.
	Decode(opt *big.Int) string
	// Number of bits needed to store opt encoding.
	BitSize() int
	// cost of using opt
	Cost(opt *big.Int) int
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

//The code for the input sources for the node.
const (
	// source from global input array
	ITypeVar = iota
	//  source from another node
	ITypeNode
	// source from a constant node of value 1
	ITypeConst
)

//The code  for output type from node.
const (
	// output not yet allocated
	OTypeNone = iota
	// output allocated to be used by a node
	OTypeNode
	// output adopted by an output
	OTypeOutput
)

//Node is  interior node Data
type Node struct {
	// the type of the input code from the IType list of costants
	ltype, rtype int
	//type of node output
	otype int
	// the index into each type array
	lindex, rindex int
	// encoded node operation
	opt *big.Int
	// store of opt cost
	optCost int
}

/*
NewDag creates a dag-function base for carrying out common  operations. opt gives
the detailed  process of converting the input pair of an interior node  into an
output. nvar gives the number of inputs to the function; nnode  is the  number
of interior nodes; nout is the number of expected outputs which should be
significantly less than nnode; nbitslookback is the number of bits  used to give
a look-back offset integer, where a look-back offset of i points to the i+1 th
object before the inner-node. Note the inputs are before the inner node and
anything before an input is regarded as a in put with a constant value of 1.

using sampleSize random samples.
*/
func NewDag(nvar, nnode, nout, nbitslookback int, opt Opt) *Dag {
	d := new(Dag)
	d.nVar = nvar
	d.nNode = nnode
	d.nOut = nout
	d.nBitsLookback = nbitslookback
	d.outNodes = make([]int, nout)
	d.INodes = make([]Node, nnode)
	d.Opt = opt
	d.optSize = d.Opt.BitSize()

	// play safe
	for i := range d.INodes {
		d.INodes[i].otype = OTypeNone
		d.INodes[i].opt = big.NewInt(0)
	}
	return d
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
func (f *FunBool) Cost(x *big.Int) *big.Int {
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

	return f.cost
}

// MaxLen returns the number of elements (bits) for the encoding.
// It is maximum number of bits in the parameter big integer which is also the
// maximum number of elements in the subset
func (d *Dag) MaxLen() int {
	return d.nNode * (2*d.nBitsLookback + d.optSize)
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

// DecodeDag requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (d *Dag) DecodeDag(z *big.Int) (s string) {
	if !d.Idecode(z) {
		s = fmt.Sprintf("DAG does not have enough outputs\n")
		return
	}
	s = fmt.Sprintf("DAG structure:\n")
	s += fmt.Sprintf("Used %d nodes\n", d.usedNodes)
	s += fmt.Sprintf("structure cost %d\n", d.structureCost)
	for i := 0; i < d.usedNodes; i++ {
		s += fmt.Sprintf("NODE %d: [ ", i)
		nd := d.INodes[i]
		// add left node input description
		switch nd.ltype {
		case ITypeVar:
			s += fmt.Sprintf("in%d ", nd.lindex)
		case ITypeNode:
			s += fmt.Sprintf("nd%d ", nd.lindex)
		case ITypeConst:
			s += fmt.Sprintf("1 ")
		}
		// add node operation description
		s += fmt.Sprintf("%s ", d.Opt.Decode(nd.opt))
		// add right node input description
		switch nd.rtype {
		case ITypeVar:
			s += fmt.Sprintf("in%d ]=>", nd.rindex)
		case ITypeNode:
			s += fmt.Sprintf("nd%d ]=>", nd.rindex)
		case ITypeConst:
			s += fmt.Sprintf("1 ]=>")
		}
		// add output type
		switch nd.otype {
		case OTypeNone:
			s += "None"
		case OTypeNode:
			s += "Node"
		case OTypeOutput:
			s += "Out"
		}
		s += fmt.Sprintln()
	}
	return

}

// Delete hints to the function to remove/replace the ith item
// it returns true if the function takes the hint
func (f *FunBool) Delete(i int) bool {
	return false
}

/*
Idecode converts  a into a corresponding DAG of interior nodes. Is used  as a
precursor to using it as a algorithm function multiple times to build up  a cost
evaluation. It returns false if it fails to do so.
*/
func (d *Dag) Idecode(a *big.Int) (ok bool) {
	ok = true
	var nodeBase int // is the binary string bit location for node data
	//populate all nodes
	for i := range d.INodes {
		nd := &d.INodes[i]
		// clear node output use
		nd.otype = OTypeNone
		// now pick up the bits to determine node operations
		opt := big.NewInt(0)
		for j := 0; j < d.optSize; j++ {
			opt.SetBit(opt, j, a.Bit(j+nodeBase))
		}
		nd.opt.Set(opt)
		// skip over opt encoding
		nodeBase += d.optSize
		if opt.Sign() == 0 {
			// treat this as an empty node that is not used
			nd.ltype = ITypeConst
			nd.lindex = 0
			nd.rtype = ITypeConst
			nd.rindex = 0
		} else {
			//l,r are look-back offsets for node input
			// make sure nodes have previous elements as input
			l := 1
			r := 1
			pwr2 := 1 //power of 2 factor to convert bits to look-back integers
			for j := 0; j < d.nBitsLookback; j++ {
				jl := j + nodeBase
				jr := jl + d.nBitsLookback
				if a.Bit(jl) == 1 {
					l += pwr2
				}
				if a.Bit(jr) == 1 {
					r += pwr2
				}
				pwr2 <<= 1
			}
			//fmt.Printf("l=%d r=%d\n", l, r)

			// process left input to node
			l = i - l   // convert to absolute node position
			if l >= 0 { // we have a previous node  as input
				nd.ltype = ITypeNode
				nd.lindex = l
				d.INodes[l].otype = OTypeNode
			} else {
				l := -l
				if l <= d.nVar { // we have a global variable input
					// notice indexing is backwards
					nd.ltype = ITypeVar
					nd.lindex = l - 1
				} else { // out of range so set to constant 1 input
					nd.ltype = ITypeConst
					nd.lindex = 0
				}
			}
			// process right input to node
			r = i - r   // convert to absolute node position
			if r >= 0 { // we have a previous node  as input
				nd.rtype = ITypeNode
				nd.rindex = r
				d.INodes[r].otype = OTypeNode
			} else {
				r := -r
				if r <= d.nVar { // we have a global variable input
					// notice indexing is backwards
					nd.rtype = ITypeVar
					nd.rindex = r - 1
				} else { // out of range so set to constant 1 input
					nd.rtype = ITypeConst
					nd.rindex = 0
				}
			}

		}
		// skip over lookback encoding
		nodeBase += 2 * d.nBitsLookback
	}
	// now allocate output to the first set of unused non empty INodes it is not
	// clear what to do if not all outputs are allocated for the moment return ok
	// as false indicating a constraint failure
	outLength := d.nOut
	nodesLength := len(d.INodes)
	o := 0
	n := 0
	d.structureCost = 0
	for ; o < outLength && n < nodesLength; n++ {
		nd := &d.INodes[n]
		if nd.otype == OTypeNone && nd.opt.Sign() != 0 {
			nd.otype = OTypeOutput
			d.outNodes[o] = n
			o++
		}
		d.structureCost += d.Opt.Cost(nd.opt)
	}
	d.usedNodes = n
	if o < outLength {
		ok = false
	}
	return
}
