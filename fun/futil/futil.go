/*
Package futil provides some usefull utilities for cost functions
*/
package futil

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
)

//W is  assumed size of word used  in big int array
const W = bits.UintSize

/*
Splitter is used to split a positive big int up into an array of big ints parts formed
from the big int. For Computational simplicity the original integer absolute
value is split up into sub slices of big.Word (assumed to be of uint); the
sign of the original big int is not used.
*/
type Splitter struct {
	offset      []int
	wordWidth   []int
	zeroingMask []big.Word
	// maximum number of bits  used including  padding to achieve word alignment.
	maxBits int
	// number of bits for each part.
	bits []int
	// required length in words.
	len int
	// array of zeros to append when needed
	zeros []big.Word
}

/* String gives a readable description of internal state mainly for diagnosis.
 */
func (sp *Splitter) String() string {
	var s string
	s += fmt.Sprintf("offsets:\n")
	for i := range sp.offset {
		s += fmt.Sprintf(" %d", sp.offset[i])
	}
	s += fmt.Sprintf("\nWidth in words:\n")
	for i := range sp.wordWidth {
		s += fmt.Sprintf(" %d", sp.wordWidth[i])

	}
	s += fmt.Sprintf("\n Zero mask:\n")
	for i := range sp.zeroingMask {
		s += fmt.Sprintf(" %x\n", sp.zeroingMask[i])
	}

	s += fmt.Sprintf("Maximum number of bits: %d\n", sp.maxBits)
	s += fmt.Sprintf("Part size in Bits:\n")
	for i := range sp.bits {
		s += fmt.Sprintf(" %d", sp.bits[i])
	}
	s += fmt.Sprintf("\nNumber of words: %d", sp.len)

	return s
}

/*
NewSplitter creates a new Splitter. Bits is an array of number of used bits for
each split element. T
*/
func NewSplitter(bits ...int) *Splitter {
	var s Splitter
	n := len(bits)
	s.bits = make([]int, n)
	s.offset = make([]int, n)
	s.wordWidth = make([]int, n)
	s.zeroingMask = make([]big.Word, n)

	origin := 0
	for i := range bits {
		s.offset[i] = origin
		x := bits[i]
		s.bits[i] = x
		k := (x / W)
		r := uint(x % W)
		if r > 0 {
			k++
			s.zeroingMask[i] = big.Word(1<<r - 1)
		} else {
			s.zeroingMask[i] = ^big.Word(0)
		}
		s.wordWidth[i] = k
		origin += k
	}
	s.len = origin
	s.maxBits = W * origin
	s.zeros = make([]big.Word, s.len)
	return &s
}

// MaxBits returns the maximum bits used by the splitter.
func (sp *Splitter) MaxBits() int {
	return sp.maxBits
}

/*Split takes a copy of the absolute value of x and splits it up into positive
big int parts in place ensuring the parts are word aligned. While doing this it
modifies the copy to match the splitting so that each part is a sub slice of
the modified copy. To do this unused padding bits are set to zero. .The parts
array is overwritten and is mapped into this copy. */
func (sp *Splitter) Split(x *big.Int, parts []*big.Int) []*big.Int {
	// do a manipulation on the local x to ensure its absolute value has enough
	// words for a direct read using Bits()
	y := big.NewInt(0)
	y.Set(x)
	words := y.Bits()
	if len(parts) != sp.len {
		parts = make([]*big.Int, sp.len)
		for i := range parts {
			parts[i] = big.NewInt(0)
		}
	}
	l := len(words)
	if l < sp.len {
		words = append(words, sp.zeros[l:sp.len]...)
	}
	for i := range sp.bits {
		begin := sp.offset[i]
		end := begin + sp.wordWidth[i]
		words[end-1] &= sp.zeroingMask[i]
		parts[i].SetBits(words[begin:end])
	}
	return parts
}

/*Join creates a positive big int from a list of word aligned parts using
splitter bit sizes and writes the result into x. it returns an error if the
parts are too large to fit or number of parts is not compatible. Also when
there is an error x remains unchanged. */
func (sp *Splitter) Join(parts []*big.Int, x *big.Int) (err error) {
	if len(parts) != sp.len {
		return fmt.Errorf("joining parts should have %d members\n ", sp.len)
	}
	// check for part fitting before transferring to x
	for i := range sp.bits {
		if sp.bits[i] < parts[i].BitLen() {
			return fmt.Errorf("part %d is too big to fit", i)
		}
	}
	//empty x to receive parts
	x.SetInt64(0)
	words := x.Bits()
	words = append(words, sp.zeros...)
	for i := range parts {
		w := parts[i].Bits()
		begin := sp.offset[i]
		end := begin + len(w)
		copy(words[begin:end], w)
	}
	x.SetBits(words)
	return nil
}

//CostValue is the data  type used to store cost values.
type CostValue interface {
	// Sets cost value using x clearing out any previous values
	Set(x interface{})
	// upates the  cost value with x by combining with previous  cost value
	// and is mainly used when Cmp returns +2 or -2.
	Update(x interface{})
	// Compares to x as a cost value
	Cmp(x interface{}) int
	// returns floatingpoint representation of number
	// of number of bits needed to represent cost as an integer.
	Fbits() float64
	// human readable value
	String() string
}

// IntCostValue is the data type used to store big integer cost values
type IntCostValue struct {
	cost *big.Int
}

// NewIntCostValue is a convenience function for generating a
// new CostValue and setting it to zero.
func NewIntCostValue() *IntCostValue {
	c := new(IntCostValue)
	c.cost = new(big.Int)
	return c
}

func (c *IntCostValue) String() string {
	return c.cost.String()
}

//Set is used to set c from x.
func (c *IntCostValue) Set(x interface{}) {
	if c1, v := x.(*IntCostValue); v {
		c.cost.Set(c1.cost)
	} else if c2, v := x.(*big.Int); v {
		c.cost.Set(c2)
	} else if c3, v := x.(int64); v {
		c.cost.SetInt64(c3)
	}

}

//Update in this case just substitutes the cost value in x
func (c *IntCostValue) Update(x interface{}) { c.Set(x) }

// Cmp compares  the cost of x with c where x is of type *IntCostValue
func (c *IntCostValue) Cmp(x interface{}) int {
	c1 := x.(*IntCostValue)
	return c.cost.Cmp(c1.cost)
}

// OpenIntCostValue returns a *IntCostValue from a x with CostValue interface //// that contains it

/*Fbits gives a floating point measure of number of bits in x that takes on non
integer values to help represent big integer size for plotting. it approximates
to the log of the big integer.
*/
func (c *IntCostValue) Fbits() float64 {
	n := c.cost.BitLen()
	if n <= 0 {
		return float64(0)
	}
	n--
	var a big.Int
	a.SetBit(&a, n, 1)
	var r big.Rat
	r.SetFrac(c.cost, &a)
	f, _ := r.Float64()
	return f + float64(n)

}

// FloatCostValue is used for floating point cost value for functions without noise
type FloatCostValue struct {
	cost float64
}

//NewFloatCostValue creates a new FloatCostValue
func NewFloatCostValue() *FloatCostValue {
	return new(FloatCostValue)
}

//Set is used to set c from x
func (c *FloatCostValue) Set(x interface{}) {
	if c1, v := x.(*FloatCostValue); v {
		*c = *c1
	} else if c2, v := x.(float64); v {
		c.cost = c2
	}
}

//Update adds the mean of x as raw data to calculate an updated stats of the cost value
func (c *FloatCostValue) Update(x interface{}) {
	c.Set(x)
}

//Cmp compares with x
func (c *FloatCostValue) Cmp(x interface{}) int {
	c1 := x.(*FloatCostValue)
	d := c.cost - c1.cost
	if d > 0 {
		return 1
	} else if d < 0 {
		return -1
	} else {
		return 0
	}
}

// Fbits scales the cost value by taking sign(x.mean)log2(1+|x.mean|)
func (c *FloatCostValue) Fbits() float64 {
	if c.cost > 0 {
		return math.Log2(1.0 + c.cost)
	}
	return -math.Log2(1 - c.cost)
}

// String is human readable value
func (c *FloatCostValue) String() string {
	return fmt.Sprintf(" %f ", c.cost)
}

//===================================================================//

/*SFloatCostValue is the data type used to store Float cost values based
  on simple statistics. Internally it maintains an mean as the cost plus the
  variance of the cost, which is used to determine if one cost is bigger than
  another; further evaluations are needed ; two costs are equivalent. Such
  things are communicated using the COMP function. As well as this it uses an
  updating weight that incorporates a forgetting process that allows for
  gradual change to the cost function itself. */
type SFloatCostValue struct {
	//treshold used in comparing values
	thres2 float64
	// mean as smoothed value
	mean float64
	// calculated variance
	variance float64
	// current memory gain set to <1
	bMem float64
	// current data gain
	lambda float64
	// limit on lambda  to ensure response to changing stats
	minLambda float64
	// sum of squared cost gains
	delta float64
	// weighted sum of squares used to calculate variance of mean
	isum float64
	// Time constant
	Tc float64
	// sigma threshold factor
	sigmaThres float64
}

// NewSFloatCostValue returns a pointer to SFloatCostValue type initialised
// with a  time constant to stats change of Tc. sigma Thres gives the number of sigmas needed to distinguish between values.
func NewSFloatCostValue(Tc, sigmaThres float64) *SFloatCostValue {

	c := new(SFloatCostValue)
	c.Tc = Tc
	c.sigmaThres = sigmaThres
	c.minLambda = 1.0 / Tc
	c.thres2 = sigmaThres * sigmaThres
	

	return c
}

//String gives human readable description
func (c *SFloatCostValue) String() string {
	return fmt.Sprintf("mean=%f variance=%f\n  TC=%f  threshold =%f ",
		c.mean, c.variance, c.Tc, c.sigmaThres)
}

//Set is used to set c from x
func (c *SFloatCostValue) Set(x interface{}) {
	if c1, v := x.(*SFloatCostValue); v {
		*c = *c1
	} else if c2, v := x.(float64); v {
		c.mean = c2
		c.variance = math.Inf(1) // play safe
		c.bMem = 1.0
		c.lambda = 1.0
		c.delta = 1.0
		c.isum = c2 * c2
	} else {
		panic("could not set cost value\n")
	}
}

//Update adds the mean of x as raw data to calculate an updated stats of the cost value
func (c *SFloatCostValue) Update(x interface{}) {
	if c.variance <= 0.0 {
		c.Set(x)
		return
	}
	c.lambda = c.lambda / (c.lambda + c.bMem)
	if c.lambda < c.minLambda {
		c.lambda = c.minLambda
		c.bMem = 1.0 - c.minLambda
	}
	c.delta = c.bMem*c.bMem*c.delta + 1.0
	c1 := x.(*SFloatCostValue)
	l1 := 1.0 - c.lambda
	l0 := c.lambda
	c.mean = l1*c.mean + l0*c1.mean
	c.isum = l1*c.isum + l0*c1.mean*c1.mean
	if dl := c.delta * l0 * l0; dl < 1.0 {
		c.variance = (c.isum - c.mean*c.mean) * dl / (1.0 - dl)
	}

}

//Cmp compares with x
func (c *SFloatCostValue) Cmp(x interface{}) int {
	c1 := x.(*SFloatCostValue)
	d := c.mean - c1.mean

	thr := c.thres2 * (c.variance + c1.variance)

	if d*d < thr {
		if c.lambda > c.minLambda {
			if c1.lambda > c.minLambda {
				if c.variance < c1.variance {
					return 2
				}
				return -2
			}
			return -2

		} else if c1.lambda > c.minLambda {
			return 2
		}
		return 0
	}

	if d > 0 {
		return 1
	} else if d < 0 {
		return -1
	} else {
		return 0
	}
}

// Fbits scales the cost value by taking sign(x.mean)log2(1+|x.mean|)
func (c *SFloatCostValue) Fbits() float64 {
	a := math.Abs(c.mean)
	fb := math.Log2(1.0 + a)
	const maxValue = 10.0
	if fb > maxValue {
		fb = maxValue
	}
	if c.mean > 0 {
		return fb
	}
	return -fb
}
