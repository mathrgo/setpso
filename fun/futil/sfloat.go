package futil

import (
	"fmt"
	"math"
	"math/big"
)

// SFloatTry is the data type used to store floating point costed try where the cost is a function of the parameter x.
type SFloatTry struct {
	x *big.Int
	TryData
	cost *SFloatCostValue
}

//Parameter reads the try value
func (t *SFloatTry) Parameter() *big.Int {
	return t.x
}

// NewSFloatTry is a convenience function for generating an
// new floating point costed try. Tc is the cost update timeconstant in iterations
func NewSFloatTry(z *big.Int, data TryData, Tc float64) *SFloatTry {
	t := new(SFloatTry)
	t.x = new(big.Int)
	t.x.Set(z)
	t.TryData = data
	t.cost = NewSFloatCostValue(Tc)

	return t
}

//Decode gives a human readable description of decoded try data
func (t *SFloatTry) Decode() string {
	return t.TryData.Decode()
}

// Cost returns a human readable cost description
func (t *SFloatTry) Cost() string {
	return t.cost.String()
}

//SetCostValue is used to set the cost value.
func (t *SFloatTry) SetCostValue(c float64) {
	t.cost.Set(c)
}

// //CostValue returns the stored cost value
// func (t *SFloatTry) CostValue() float64 {
// 	return t.cost
// }

//Cmp compares  the cost of t with s where t is of type *SFloatTry
func (t *SFloatTry) Cmp(s *SFloatTry, mode CmpMode) float64 {
	return t.cost.Cmp(s.cost, mode)
}

//Data returns the decoded data
func (t *SFloatTry) Data() TryData {
	return t.TryData
}

/*Fbits gives a floating point measure of number of bits in the
cost that takes on non integer values to help represent
big integer cost size for plotting. it approximates
to the log of the big integer.
*/
func (t *SFloatTry) Fbits() float64 {
	cost := t.cost.mean
	if cost > 0 {
		return math.Log2(1.0 + cost)
	}
	return -math.Log2(1 - cost)

}

//SFloatFun is the interface for big int costed function
type SFloatFun interface {
	Fun
	// calculates the cost of the try using the decoded data, returning  the cost
	Cost(data TryData) float64
}

// SFloatFunStub uses SFloatFun interface to create the setpso.Fun interface
type SFloatFunStub struct {
	SFloatFun
	// initial time  constant of try cost updates
	Tc float64
	//comparison margin in sigmas
	SigmaMargin float64
}

//Fun retrieves the internal cost function
func (f *SFloatFunStub) Fun() SFloatFun { return f.SFloatFun }

//NewSFloatFunStub creates an instance of the SFloatFunStub ready for use as the interface setpso.Fun. Tc is the initial try cost update time constant.
func NewSFloatFunStub(f SFloatFun, Tc, SigmaMargin float64) *SFloatFunStub {
	stub := new(SFloatFunStub)
	stub.SFloatFun = f
	stub.Tc = Tc
	stub.SigmaMargin = SigmaMargin * SigmaMargin
	return stub
}

//NewTry creates a try as an SFloatTry
func (f *SFloatFunStub) NewTry() Try {
	try := NewSFloatTry(f.DefaultParam(), f.CreateData(), f.Tc)
	f.IDecode(try.TryData, try.Parameter())
	try.cost.Set(f.Cost(try.TryData))
	//measure cost again to get variance
	//try.cost.Update(f.Cost(try.TryData))
	return try
}

//SetTry sets try  to a new parameter z
func (f *SFloatFunStub) SetTry(t Try, z *big.Int) {
	try := t.(*SFloatTry)
	try.x.Set(z)
	//fmt.Printf("Parameter=%v\n",try.Parameter())
	f.IDecode(try.TryData, try.Parameter())
	try.cost.Set(f.Cost(try.TryData))
	//measure cost again to get variance
	try.cost.Update(f.Cost(try.TryData))
}

//Copy copies src to dest
func (f *SFloatFunStub) Copy(dest, src Try) {
	d := dest.(*SFloatTry)
	s := src.(*SFloatTry)
	d.x.Set(s.x)
	f.CopyData(d.TryData, s.TryData)
	d.cost.Copy(s.cost)
}

//UpdateCost recalculates the try cost
func (f *SFloatFunStub) UpdateCost(t Try) {
	try := t.(*SFloatTry)
	try.cost.Update(f.Cost(try.TryData))
}

//Cmp compares the tries
func (f *SFloatFunStub) Cmp(x, y Try, mode CmpMode) float64 {
	s := x.(*SFloatTry)
	t := y.(*SFloatTry)

	result:=t.Cmp(s, mode)
	if mode == TriesMode {
		result = result/f.SigmaMargin
	}
	return result
}

// ToConstraint uses the previous try pre and the updating hint parameter
// to attempt to produce an update to pre which satisfies
// solution constraints it returns valid = True if succeeds, otherwise pre remains un changed and returns false
func (f *SFloatFunStub) ToConstraint(pre Try, hint *big.Int) bool {
	p := pre.(*SFloatTry)
	if f.Constraint(p.TryData, hint) {
		f.SetTry(pre, hint)
		return true
	}
	return false
}

//===================================================================//

/*SFloatCostValue is the data type used to store Float cost values based
  on simple statistics.
  Internally it maintains an mean as the cost plus a count of  number of times the mean has been sampled as well as a success count  , which is used to determine how two costs compare.
  Such things are communicated using the Cmp function.
  As well as this it uses an updating weight that incorporates a forgetting process that allows for gradual change to the cost function itself. */
type SFloatCostValue struct {
	// mean as smoothed value of cost
	mean float64
	// remembering gain
	alpha float64
	//effective sum of costs
	costSum float64
	//effective number of updates
	updateSum float64
	//effective number of comparisons
	compSum float64
	//effective number of comparisons leading to better cost
	compSuccessSum float64
	// success score as probability-0.5
	epsilon float64
	// Time constant
	Tc float64
}

// Copy takes a copy of c1
func (c *SFloatCostValue) Copy(c1 *SFloatCostValue) {
	// play safe by using explicit copy
	c.mean = c1.mean
	c.alpha = c1.alpha
	c.costSum = c1.costSum
	c.updateSum = c1.updateSum
	c.compSum = c1.compSum
	c.compSuccessSum = c1.compSuccessSum
	c.epsilon = c1.epsilon
	c.Tc = c1.Tc
}

// NewSFloatCostValue returns a pointer to SFloatCostValue type initialized
// with a  time constant to stats change of Tc.
func NewSFloatCostValue(Tc float64) *SFloatCostValue {

	c := new(SFloatCostValue)
	c.Tc = Tc
	c.alpha = 1.0 - 1.0/Tc

	return c
}

//String gives human readable description
func (c *SFloatCostValue) String() string {
	return fmt.Sprintf("mean=%f updates=%f success=%f comps=%f TC=%f  \n ",
		c.mean, c.updateSum, c.epsilon, c.compSum, c.Tc)
}

//Set is used to set c from x
func (c *SFloatCostValue) Set(x float64) {
	c.mean = x
	c.updateSum = 1.0
}

//Update adds the mean of x as raw data to calculate an updated stats of the cost value
func (c *SFloatCostValue) Update(x float64) {
	c.updateSum *= c.alpha
	c.updateSum++
	c.costSum *= c.alpha
	c.costSum += x
	c.mean = c.costSum / c.updateSum
}

//Cmp compares with x
func (c *SFloatCostValue) Cmp(c1 *SFloatCostValue, mode CmpMode) float64 {
	d := c1.mean - c.mean

	switch mode {
	case CostMode:
		// wipe out tries history
		c.compSum =0
		c.compSuccessSum=0 
		// limit comparison certainty to indicate that it is not deterministic
		if d > 0 {
			return 0.5
		}
		return -0.5
	case TriesMode:
		c.compSum *= c.alpha
		c.compSum++
		if d > 0 {
			c.compSuccessSum *= c.alpha
			c.compSuccessSum++
		}
		c.epsilon = 0.5 * (2*c.compSuccessSum - c.compSum) / (c.compSum + 2)
		epsilon2 := c.epsilon * c.epsilon
		s := c.compSum * epsilon2 / (0.25 - epsilon2)
		if c.epsilon >= 0 {
			return s
		}
		return -s
	}
	return d //fail safe?
}

// Fbits scales the cost value by taking sign(x.mean)log2(1+|x.mean|)
func (c *SFloatCostValue) Fbits() float64 {
	a := math.Abs(c.mean)
	fb := math.Log(1.0 + a)
	const maxValue = 10.0
	if fb > maxValue {
		fb = maxValue
	}
	if c.mean > 0 {
		return fb
	}
	return -fb
}
