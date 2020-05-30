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
	Opt(l, r float64, opt *big.Int) float64
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
func (o *Opt2Float) Opt(l, r float64, opt *big.Int) float64 {
	var f float64
	switch opt.Int64() {
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
	Float(int) float64
	// maximum number of bits to represent the integer
	BitSize() int
}

//Int2FloatList uses a list of values to give the int to float index where the index is obtained taking modulus of the index length. Typicall the length of the list is a power of 2.
type Int2FloatList struct {
	bits int
	list []float64
}

//NewInt2FloatList creates an Int2floatList from a list of values.
func NewInt2FloatList(a ...float64) *Int2FloatList {
	f := new(Int2FloatList)
	f.list = make([]float64, len(a))
	copy(f.list, a)
	f.bits = bits.Len(uint(len(a)))
	return f
}

//BitSize gives the number of bits needed for the index.
func (f *Int2FloatList) BitSize() int {
	return f.bits
}

//Float returns the corresponding float value where the index od modded with the list length to avoid over flow.
func (f *Int2FloatList) Float(index int) float64 {
	return f.list[index%len(f.list)]
}

//=====================================================

//Int2FloatRange converts bits to a range of floats
type Int2FloatRange struct {
	bits        int
	gain float64
	begin, end  float64
}

//NewInt2FloatRange creates an Int2floatRange  from number of bits to use  and the corresponding floating point range  .
func NewInt2FloatRange(nbits int, begin, end float64) *Int2FloatRange {
	f := new(Int2FloatRange)
	f.bits = nbits
	f.begin = begin
	f.end = end
	f.gain = (f.end-f.begin)/float64(int(1)<<uint(nbits)) 


	return f
}

//BitSize gives the number of bits needed for the index.
func (f *Int2FloatRange) BitSize() int {
	return f.bits
}

//Float returns the corresponding float value 
func (f *Int2FloatRange) Float(index int) float64 {
	return f.begin+f.gain*float64(index)
}
//===========================================

/*
OptMorphFloat encodes the Float operation  in a uniform way so the nodes can morph easily between algabraic and power law values. The operator has at first sight the strange form:
	z=h(il;x)+ h(ir;y)
operating on the pair (x,y) to give z at a node.
the binary string representing the operation is il|ir where il and ir sub binary sequences for left and right parts of the input to the node.
	h(sc|jc|sp|jp,x)=
	(-1)^sc*(sign(x))^sp*C(jc)*|x|^P(jp)
where sc,sp are 0 or 1 ;jc,jp  short non negative integers in the range 0<= jc<2^nc , 0<= jp<2^np that are mapped by the functions C and P to floating point values.
nc,np are natural numbers giving the number of bits used to represent jc, jp. C and P are of type Int2Float.

this is supprisingly expressive for instance a ratio can be expressed usig just 3 nodes.

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
func (o *OptMorphFloat) Cost(opt *big.Int) int {
	return 1
}

// About gives description of operation type
func (o *OptMorphFloat) About() string {
	s := " general morphing float operation"
	return s
}

// Decode gives a human readable discriptin of opt.
func (o *OptMorphFloat) Decode(opt *big.Int) string {
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
	return 1 + o.tC.BitSize() + o.tP.BitSize()
}

//IDecode splits the opt into its consituent parts
func (o *OptMorphFloat) IDecode(z *big.Int) {
	scl := z.Bit(0) > 0
	b := 1
	index := 0
	for i := 0; i < o.tC.BitSize(); i++ {
		index *= 2
		if z.Bit(b) > 0 {
			index++
		}
		b++
	}
	if scl {
		o.cl = -o.tC.Float(index)
	} else {
		o.cl = o.tC.Float(index)
	}
	o.spl = z.Bit(b) > 0
	b++
	index = 0
	for i := 0; i < o.tP.BitSize(); i++ {
		index *= 2
		if z.Bit(b) > 0 {
			index++
		}
		b++
	}
	o.pl = o.tP.Float(index)

	scr := z.Bit(b) > 0
	b++
	index = 0
	for i := 0; i < o.tC.BitSize(); i++ {
		index *= 2
		if z.Bit(b) > 0 {
			index++
		}
		b++
	}
	if scr {
		o.cr = -o.tC.Float(index)
	} else {
		o.cr = o.tC.Float(index)
	}
	o.spr = z.Bit(b) > 0
	b++
	index = 0
	for i := 0; i < o.tP.BitSize(); i++ {
		index *= 2
		if z.Bit(b) > 0 {
			index++
		}
		b++
	}
	o.pr = o.tP.Float(index)
}

// Opt is function to generate a node output. l,r are the left and right inputs
// and opt is the encoded operation to be carried out on l,r to return the
// result.
func (o *OptMorphFloat) Opt(l, r float64, opt *big.Int) float64 {

	fl := o.cl * math.Pow(math.Abs(l), o.pl)
	if o.spl && l < 0.0 {
		fl = -fl
	}
	fr := o.cl * math.Pow(math.Abs(r), o.pr)
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
	// store of cost value for optimiser
	costValue *futil.SFloatCostValue

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
	// treshold in sigmas for sinificant comparison
	sigmaThres float64
	// input buffer of samples
	inbuf []float64
	// output buffer of samples
	outbuf []float64
	// life of buffer in cycles
	sampleLife int
	// life counter
	count int
}

// NewFunFloat returns a new *FunFloat ready to be used.
func NewFunFloat(nnode, nbitslookback int, opt OptFloat, sizeCostFactor float64,
	sampler SamplerFloat, sampleSize int,
	rnd *rand.Rand, Tc, sigmaThres float64, sampleLife int) *FunFloat {
	nvar := sampler.InputSize()
	nout := sampler.OutputSize()

	f := FunFloat{NewDag(nvar, nnode, nout, nbitslookback, opt),
		make([]float64, nnode),
		make([]float64, nvar),
		make([]float64, nout),
		math.Inf(1),
		futil.NewSFloatCostValue(Tc, sigmaThres),
		0.0,
		opt,
		sampler,
		sizeCostFactor,
		sampleSize,
		rnd,
		Tc,
		sigmaThres,
		make([]float64, nvar*sampleSize),
		make([]float64, nout*sampleSize),
		sampleLife,
		0,
	}
	return &f
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
func (f *FunFloat) Cost(x *big.Int) futil.CostValue {
	f.Idecode(x) // assume satisfies constraints
	//fmt.Printf("x=%v %s\n",x,f.Decode(x))
	nvar := f.sampler.InputSize()
	nout := f.sampler.OutputSize()
	jin := 0
	jout := 0
	f.difCost = 0.0
	f.cost = float64(f.structureCost) * f.sizeCostFactor
	f.count--
	if f.count <= 0 {
		f.count = f.sampleLife
		jin = 0
		jout = 0
		for i := 0; i < f.sampleSize; i++ {
			f.sampler.Sample(f.inbuf[jin:jin+nvar], f.outbuf[jout:jout+nout], f.rnd)
			jin += nvar
			jout += nout
		}
	}
	jin = 0
	jout = 0
	for j := 0; j < f.sampleSize; j++ {

		f.input = f.inbuf[jin : jin+nvar]
		f.output = f.outbuf[jout : jout+nout]
		jin += nvar
		jout += nout
		//fmt.Printf("in: %v out: %v \n", f.input, f.output)

		// evaluate dag nodes using f.input from sample
		for i := 0; i < f.usedNodes; i++ {
			nd := &f.INodes[i]
			var l float64
			switch nd.ltype {
			case ITypeVar:
				l = f.input[nd.lindex]
			case ITypeNode:
				l = f.nodeValues[nd.lindex]
			case ITypeConst:
				l = 1.0
			}
			var r float64
			switch nd.rtype {
			case ITypeVar:
				r = f.input[nd.rindex]
			case ITypeNode:
				r = f.nodeValues[nd.rindex]
			case ITypeConst:
				r = 1.0
			}
			f.nodeValues[i] = f.opt.Opt(l, r, nd.opt)
		}
		// calculate missmatch with output
		c := 0.0
		for i := 0; i < nout; i++ {
			dif := f.nodeValues[f.outNodes[i]] - f.output[i]
			c += dif * dif
		}
		f.difCost += c
	}
	//fmt.Printf("difcost= %f\n",f.difCost)
	f.difCost /= float64(f.sampleSize)
	f.cost += f.difCost
	f.costValue.Set(f.cost)
	//fmt.Printf("costValue %v \n", f.costValue)
	return f.costValue
}

// About string gives a description of the cost function
func (f *FunFloat) About() string {
	s := fmt.Sprintf("Dag using operation: %s\n", f.opt.About())
	s += fmt.Sprintf("Sampler: %s\n ", f.sampler.About())
	s += fmt.Sprintf("sample size %d\n", f.sampleSize)
	s += fmt.Sprintf("Dag node cost factor %f", f.sizeCostFactor)
	return s
}

//Decode gives human radable description of encoding
func (f *FunFloat) Decode(z *big.Int) string {
	s := f.DecodeDag(z)
	return s
}

//ToConstraint attempts to  give a constraint satisfying hint that matches the hint;
// pre is the previous constraint satisfying version to hint, which
// should not be changed. It returns True if it succeeds.
func (f *FunFloat) ToConstraint(pre, hint *big.Int) bool {
	ok := f.Idecode(hint)
	return ok
}

// Delete hints to the function to remove/replace the ith item
// it returns true if the function takes the hint
func (f *FunFloat) Delete(i int) bool {
	return false
}

//NewCostValue creates zero cost value
func (f *FunFloat) NewCostValue() futil.CostValue {
	return futil.NewSFloatCostValue(f.Tc, f.sigmaThres)
}
