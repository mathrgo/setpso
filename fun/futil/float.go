package futil

import (
	"fmt"
	"math"
	"math/big"
)

// FloatTry is the data type used to store floating point costed try where the cost is a function of the parameter x.
type FloatTry struct {
	x *big.Int
	TryData
	cost float64
}

//Parameter reads the try value
func (t *FloatTry) Parameter() *big.Int {
	return t.x
}

// NewFloatTry is a convenience function for generating an
// new floating point costed try.
func NewFloatTry(z *big.Int, data TryData) *FloatTry {
	t := new(FloatTry)
	t.x = new(big.Int)
	t.x.Set(z)
	t.TryData = data

	return t
}

//Decode gives a human readable description of decoded try data
func (t *FloatTry) Decode() string {
	return t.TryData.Decode()
}

// Cost returns a human readable cost description
func (t *FloatTry) Cost() string {
	return fmt.Sprintf(" %f", t.cost)
}

//SetCostValue is used to set the cost value.
func (t *FloatTry) SetCostValue(c float64) {
	t.cost = c
}

//CostValue returns the stored cost value
func (t *FloatTry) CostValue() float64 {
	return t.cost
}

//Cmp compares  the cost of t with s where t is of type *FloatTry
func (t *FloatTry) Cmp(s *FloatTry) float64 {
	d := t.cost - s.cost
	if d > 0 {
		return 1
	} else if d < 0 {
		return -1
	} else {
		return 0
	}
}

//Data returns the decoded data
func (t *FloatTry) Data() TryData {
	return t.TryData
}

/*Fbits gives a floating point measure of number of bits in the
cost that takes on non integer values to help represent
big integer cost size for plotting. it approximates
to the log of the big integer.
*/
func (t *FloatTry) Fbits() float64 {
	if t.cost > 0 {
		return math.Log2(1.0 + t.cost)
	}
	return -math.Log2(1 - t.cost)

}

//FloatFun is the interface for big int costed function
type FloatFun interface {
	Fun
	// calculates the cost of the try using the decoded data, returning  the cost
	Cost(data TryData) float64
}

// FloatFunStub uses FloatFun interface to create the setpso.Fun interface
type FloatFunStub struct {
	FloatFun
}

//Fun retrieves the internal cost function
func (f *FloatFunStub) Fun() FloatFun { return f.FloatFun }

//NewFloatFunStub creates an instance of the FloatFunStub ready for use as the interface setpso.Fun
func NewFloatFunStub(f FloatFun) *FloatFunStub {
	stub := new(FloatFunStub)
	stub.FloatFun = f
	return stub
}

//NewTry creates a try as an FloatTry
func (f *FloatFunStub) NewTry() Try {
	try := NewFloatTry(f.DefaultParam(), f.CreateData())
	f.IDecode(try.TryData, try.Parameter())
	try.cost = f.Cost(try.TryData)
	return try
}

//SetTry sets try  to a new parameter z
func (f *FloatFunStub) SetTry(t Try, z *big.Int) {
	try := t.(*FloatTry)
	try.x.Set(z)
	//fmt.Printf("Parameter=%v\n",try.Parameter())
	f.IDecode(try.TryData, try.Parameter())
	try.cost = f.Cost(try.TryData)
}

//Copy copies src to dest
func (f *FloatFunStub) Copy(dest, src Try) {
	d := dest.(*FloatTry)
	s := src.(*FloatTry)
	d.x.Set(s.x)
	f.CopyData(d.TryData, s.TryData)
	d.cost = s.cost
}

//UpdateCost recalculates the try cost
func (f *FloatFunStub) UpdateCost(t Try) {
	try := t.(*FloatTry)
	try.cost = f.Cost(try.TryData)
}

//Cmp compares the tries
func (f *FloatFunStub) Cmp(x, y Try, mode CmpMode) float64 {
	s := x.(*FloatTry)
	t := y.(*FloatTry)
	return s.Cmp(t)
}

// ToConstraint uses the previous try pre and the updating hint parameter
// to attempt to produce an update to pre which satisfies
// solution constraints it returns valid = True if succeeds, otherwise pre remains un changed and returns false
func (f *FloatFunStub) ToConstraint(pre Try, hint *big.Int) bool {
	p := pre.(*FloatTry)
	if f.Constraint(p.TryData, hint) {
		f.SetTry(pre, hint)
		return true
	}
	return false
}
