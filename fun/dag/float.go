package dag

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"math/rand"

	"github.com/mathrgo/setpso/fun/futil"
)

/*
OptFloat is  interface for interior node operation using float operations . l and r are the left and right
inputs to the node, opt is a, positive integer,  operations code
*/
type OptFloat interface {
	Opt
	// function to generate a node output l,r are the left and right inputs and opt
	// is the encoded operation to be carried out on l,r to return the result.
	Opt(l, r float64, opt uint) float64
}

/*
Opt2Float encodes the Float operation as a choice betwee '+','-','/','^' operations
*/
type Opt2Float struct {
	// symbols for representing operations in a human readable form
	symbol [4]string
	// cost  of using node table
	NodeCost [4]int
}

/*
NewOpt2Float creates a Opt2Float and inserts default node cost values
*/
func NewOpt2Float() *Opt2Float {
	var o Opt2Float
	o.symbol = [4]string{"+", "-", "/", "^"}
	// set up default node use cost
	o.NodeCost = [4]int{1, 1, 1, 1}
	return &o
}

// Cost gives the cost of using a node
func (o *Opt2Float) Cost(opt *big.Int) int {
	return o.NodeCost[int(opt.Int64())]
}

// About gives description of operation type
func (o *Opt2Float) About() string {
	s := " general float operation using power and divide"
	return s
}

// Decode gives a human readable discriptin of opt.
func (o *Opt2Float) Decode(opt *big.Int) string {
	return o.symbol[int(opt.Int64())]
}

// BitSize returns Number of bits needed to store opt encoding.
func (o *Opt2Float) BitSize() int {
	return 2
}

// Opt is function to generate a node output. l,r are the left and right inputs
// and opt is the encoded operation to be carried out on l,r to return the
// result.
func (o *Opt2Float) Opt(l, r float64, opt int) float64 {
	var f float64
	switch opt {
	case 0:
		f = l + r
	case 1:
		f = l - r
	case 3:
		if r == 0 {
			if l < 0 {
				f = math.Inf(-1)
			} else {
				f = math.Inf(1)
			}
		} else if r == math.Inf(-1) {
			f = 0.0
		} else if r == math.Inf(1) {
			f = 0.0
		} else {
			f = l / r
		}
	case 4:
		f = math.Pow(math.Abs(l), math.Abs(r))
	}
	return f
}

//===========================================

//Int2Float gives a mapping of an indexing  integer to a float.
type Int2Float interface {
	// the mapping
	Float(uint) float64
	// maximum number of least significant bits of the integer used
	BitSize() uint
}

//Int2FloatList uses a list of values to give the int to float index where the index is obtained taking modulus of the index length. Typicall the length of the list is a power of 2.
type Int2FloatList struct {
	bits uint
	list []float64
	mask uint
}

//NewInt2FloatList creates an Int2floatList from a list of values.
func NewInt2FloatList(a ...float64) *Int2FloatList {
	f := new(Int2FloatList)
	f.list = make([]float64, len(a))
	copy(f.list, a)
	f.bits = uint(bits.Len(uint(len(a))))
	f.mask = 1<<f.bits - 1
	return f
}

//BitSize gives the number of bits needed for the index.
func (f *Int2FloatList) BitSize() uint {
	return f.bits
}

//Float returns the corresponding float value where the index od modded with the list length to avoid over flow.
func (f *Int2FloatList) Float(index uint) float64 {
	index &= f.mask
	return f.list[index%uint(len(f.list))]
}

//=====================================================

//Int2FloatRange converts bits to a range of floats
type Int2FloatRange struct {
	bits       uint
	gain       float64
	begin, end float64
	mask       uint
}

//NewInt2FloatRange creates an Int2floatRange  from number of bits to use  and the corresponding floating point range  .
func NewInt2FloatRange(nbits int, begin, end float64) *Int2FloatRange {
	f := new(Int2FloatRange)
	f.bits = uint(nbits)
	f.begin = begin
	f.end = end
	f.gain = (f.end - f.begin) / float64(int(1)<<uint(nbits))
	f.mask = 1<<f.bits - 1

	return f
}

//BitSize gives the number of bits needed for the index.
func (f *Int2FloatRange) BitSize() uint {
	return f.bits
}

//Float returns the corresponding float value
func (f *Int2FloatRange) Float(index uint) float64 {
	index &= f.mask
	return f.begin + f.gain*float64(index)
}

//===========================================

/*
OptMorphFloat encodes the Float operation  in a uniform way so the nodes can morph easily between algebraic and power law values. The operator has at first sight the strange form:
	z=h(il;x)+ h(ir;y)
operating on the pair (x,y) to give z at a node.
the binary string representing the operation is il|ir where il and ir sub binary sequences for left and right parts of the input to the node.
	h(sc|jc|sp|jp,x)=
	(-1)^sc*(sign(x))^sp*C(jc)*|x|^P(jp)
where sc,sp are 0 or 1 ;jc,jp  short non negative integers in the range 0<= jc<2^nc , 0<= jp<2^np that are mapped by the functions C and P to floating point values.
nc,np are natural numbers giving the number of bits used to represent jc, jp. C and P are of type Int2Float.

this is surprisingly expressive for instance a ratio can be expressed using just 3 nodes.

*/
type OptMorphFloat struct {
	// decoders
	tC, tP Int2Float
	// power sign bits
	spl, spr bool
	// coefficient parts
	cl, cr float64
	// power parts
	pl, pr float64
}

/*
NewOptMorphFloat creates a OptMorphFloat and inserts default node cost values
*/
func NewOptMorphFloat(C, P Int2Float) *OptMorphFloat {
	var o OptMorphFloat
	o.tC = C
	o.tP = P
	return &o
}

// Cost gives the cost of using a node
func (o *OptMorphFloat) Cost(opt uint) int {
	return 1
}

// About gives description of operation type
func (o *OptMorphFloat) About() string {
	s := " general morphing float operation"
	return s
}

// Decode gives a human readable discriptin of opt.
func (o *OptMorphFloat) Decode(opt uint) string {
	o.IDecode(opt)
	var s string
	if o.spl {
		s += fmt.Sprintf(" %f*x^%f + ", o.cl, o.pl)
	} else {
		s += fmt.Sprintf(" %f*|x|^%f + ", o.cl, o.pl)
	}
	if o.spr {
		s += fmt.Sprintf(" %f*y^%f ", o.cr, o.pr)
	} else {
		s += fmt.Sprintf(" %f*|y|^%f ", o.cr, o.pr)
	}

	return s
}

// BitSize returns Number of bits needed to store opt encoding.
func (o *OptMorphFloat) BitSize() int {
	return int(2 + o.tC.BitSize() + o.tP.BitSize())*2
}

//IDecode splits the opt into its constituent parts
func (o *OptMorphFloat) IDecode(z uint) {
	// pick out sign of cl
	scl := z&1 > 0
	z >>= 1
	// compute cl with sign
	if scl {
		o.cl = -o.tC.Float(z)
	} else {
		o.cl = o.tC.Float(z)
	}
	z >>= o.tC.BitSize()
	// get sign of pl
	o.spl = z&1 > 0
	// compute pl
	o.pl = o.tP.Float(z)
	z >>= o.tP.BitSize()
	// get sigh of cr
	scr := z&1 > 0
	z >>= 1
	// compute cr with sign
	if scr {
		o.cr = -o.tC.Float(z)
	} else {
		o.cr = o.tC.Float(z)
	}
	z >>= o.tC.BitSize()
	// get sign of pr
	o.spr = z&1 > 0
	z >>= 1
	// compute pr
	o.pr = o.tP.Float(z)
}

// Opt is function to generate a node output. l,r are the left and right inputs
// and opt is the encoded operation to be carried out on l,r to return the
// result.
func (o *OptMorphFloat) Opt(l, r float64, opt uint) float64 {
	o.IDecode(opt)// would like to not do this  every time
	fl := o.cl * math.Pow(math.Abs(l), o.pl)
	if o.spl && l < 0.0 {
		fl = -fl
	}
	fr := o.cr * math.Pow(math.Abs(r), o.pr)
	if o.spr && r < 0.0 {
		fr = -fr
	}
	return fl + fr
}

//===========================================

/*SamplerFloat is an interface for sampling inputs and required output results. Where x is the input array;
y is the corresponding required output array of bits given by the sampler; both arrays stored as big integers; rnd is the random number generator
used to create samples.
*/
type SamplerFloat interface {
	Sample(x, y []float64, rnd *rand.Rand)
	InputSize() int
	OutputSize() int
	About() string
}

/*
FunFloat is the cost-function for evaluating boolean DAG encoded as a (positive) big integer.
*/
type FunFloat struct {
	// temporary store of dag decoder
	*Dag
	// temporary store of dag values
	nodeValues []float64
	//temporary store of input
	input []float64
	//temporary store of output
	output []float64
	// temporary store of cost value
	cost float64

	// temporary store of difference between output and required output
	// cost
	difCost float64
	// node operator
	opt OptFloat
	// sample generator
	sampler SamplerFloat

	//factor weight to include INode usage cost
	sizeCostFactor float64
	// random sample size for evaluating output missmatch cost
	sampleSize int
	// random number generator
	rnd *rand.Rand
	// Timeconstant for cost evaluation
	Tc float64
	// treshold in sigmas for significant comparison
	sigmaThres float64
}

//FloatTry gives the try structure to use
type FloatTry = futil.SFloatTry

//FloatFunStub gives interface to setpso
type FloatFunStub = futil.SFloatFunStub

// NewFunFloat returns a new *FunFloat ready to be used.
func NewFunFloat(nnode, nbitslookback int, opt OptFloat, sizeCostFactor float64,
	sampler SamplerFloat, sampleSize int,
	rnd *rand.Rand, Tc, sigmaThres float64) *FloatFunStub {
	nvar := sampler.InputSize()
	nout := sampler.OutputSize()
	var f FunFloat
	f.Dag = NewDag(nvar, nnode, nout, nbitslookback, opt)
	f.nodeValues = make([]float64, nnode)
	f.input = make([]float64, nvar)
	f.output = make([]float64, nout)
	f.cost = math.Inf(1)
	f.difCost = 0.0
	f.opt = opt
	f.sampler = sampler
	f.sizeCostFactor = sizeCostFactor
	f.sampleSize = sampleSize
	f.rnd = rnd
	f.Tc = Tc
	f.sigmaThres = sigmaThres

	return futil.NewSFloatFunStub(&f, Tc, sigmaThres)
}

// SetSizeCostFactor sets factor weight to include INode usage cost
func (f *FunFloat) SetSizeCostFactor(sizeCostFactor float64) {
	f.sizeCostFactor = sizeCostFactor
}

//SizeCostFactor returns factor weight to include INode usage cost
func (f *FunFloat) SizeCostFactor() float64 {
	return f.sizeCostFactor
}

/*
Cost evaluates cost, where a lower cost is better. In this case
where

	cost = node usage cost * sizeCostFactor
	       + sum of absolute value of output mismatches

using random samples.
*/
func (f *FunFloat) Cost(data TryData) (cost float64) {
	d := data.(*DTryData)
	f.difCost = 0.0
	sCost := float64(d.structureCost)
	cost = sCost * f.sizeCostFactor
	nout := f.sampler.OutputSize()
	for j := 0; j < f.sampleSize; j++ {
		f.sampler.Sample(f.input, f.output, f.rnd)

		// evaluate dag nodes using f.input from sample
		for i := 0; i < d.usedNodes; i++ {
			nd := &d.INodes[i]
			var l float64
			switch nd.ltype {
			case ITypeVar:
				l = f.input[nd.lindex]
			case ITypeNode:
				l = f.nodeValues[nd.lindex]
			case ITypeConst:
				l = 0
			}
			var r float64
			switch nd.rtype {
			case ITypeVar:
				r = f.input[nd.rindex]
			case ITypeNode:
				r = f.nodeValues[nd.rindex]
			case ITypeConst:
				r = 0
			}
			f.nodeValues[i] = f.opt.Opt(l, r, nd.opt)
		}
		// calculate missmatch with output
		c := 0.0
		for i := 0; i < nout; i++ {
			dif := f.nodeValues[d.outNodes[i]] - f.output[i]
			c += dif * dif
		}
		f.difCost += c
	}
	//fmt.Printf("difCost= %f\n",f.difCost)
	f.difCost /= float64(f.sampleSize)
	cost += f.difCost
	return

}

//DefaultParam gives a default that satisfies constraints
func (f *FunFloat) DefaultParam() *big.Int {
	return big.NewInt(0)
}

// About string gives a description of the cost function
func (f *FunFloat) About() string {
	s := fmt.Sprintf("Dag using operation: %s\n", f.opt.About())
	s += fmt.Sprintf("Sampler: %s\n ", f.sampler.About())
	s += fmt.Sprintf("sample size %d\n", f.sampleSize)
	s += fmt.Sprintf("Dag node cost factor %f", f.sizeCostFactor)
	return s
}

// Delete hints to the function to remove/replace the ith item
// it returns true if the function takes the hint
func (f *FunFloat) Delete(i int) bool {
	return false
}
