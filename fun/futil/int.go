package futil

import (
	"math/big"
)

// IntTry is the data type used to store big integer costed try.
type IntTry struct {
	x *big.Int
	TryData
	cost *big.Int
}

//Parameter reads the try value
func (t *IntTry) Parameter() *big.Int {
	return t.x
}

// NewIntTry is a convenience function for generating an
// new big int costed try.
func NewIntTry(z *big.Int, data TryData) *IntTry {
	t := new(IntTry)
	t.cost = new(big.Int)
	t.x = new(big.Int)
	t.x.Set(z)
	t.TryData = data

	return t
}

//Decode gives a human readable description of decoded try data
func (t *IntTry) Decode() string {
	return t.TryData.Decode()
}

// Cost returns a human readable cost description
func (t *IntTry) Cost() string {
	return t.cost.String()
}

//SetCostValue is used to set the cost value.
func (t *IntTry) SetCostValue(c *big.Int) {
	t.cost.Set(c)
}

//CostValue returns the stored cost value
func (t *IntTry) CostValue() *big.Int {
	return t.cost
}

//Cmp compares  the cost of x with c where x is of type *IntTry
func (t *IntTry) Cmp(s *IntTry) float64 {
	if t.cost.Cmp(s.cost)<0 {
		return -1.0
	}
	return 1.0
}

//Data returns the decoded data
func (t *IntTry) Data() TryData {
	return t.TryData
}

/*Fbits gives a floating point measure of number of bits in the
cost that takes on non integer values to help represent
big integer cost size for plotting. it approximates
to the log of the big integer.
*/
func (t *IntTry) Fbits() float64 {
	n := t.cost.BitLen()
	if n <= 0 {
		return float64(0)
	}
	n--
	var a big.Int
	a.SetBit(&a, n, 1)
	var r big.Rat
	r.SetFrac(t.cost, &a)
	f, _ := r.Float64()
	return f + float64(n)

}

//IntFun is the interface for big int costed function
type IntFun interface {
	Fun
	// calculates the cost of the try using the decoded data, returning  the result in cost
	Cost(data TryData, cost *big.Int)
}

// IntFunStub uses IntFun interface to create the setpso.Fun interface
type IntFunStub struct {
	IntFun
	tempCost *big.Int
}

//NewIntFunStub creates an instance of the IntFunStub ready for use as the interface setpso.Fun
func NewIntFunStub(f IntFun) *IntFunStub {
	stub := new(IntFunStub)
	stub.IntFun = f
	stub.tempCost = new(big.Int)
	return stub
}

//NewTry creates a try as an IntTry
func (f *IntFunStub) NewTry() Try {
	try := NewIntTry(f.DefaultParam(), f.CreateData())
	f.IDecode(try.TryData, try.Parameter())
	f.Cost(try.TryData, f.tempCost)
	try.SetCostValue(f.tempCost)
	return try
}

//SetTry sets try  to a new parameter z
func (f *IntFunStub) SetTry(t Try, z *big.Int) {
	try := t.(*IntTry)
	try.x.Set(z)
	//fmt.Printf("Parameter=%v\n",try.Parameter())
	f.IDecode(try.TryData, try.Parameter())
	f.Cost(try.TryData, f.tempCost)
	try.SetCostValue(f.tempCost)
}

//Copy copies src to dest
func (f *IntFunStub) Copy(dest, src Try) {
	d := dest.(*IntTry)
	s := src.(*IntTry)
	d.x.Set(s.x)
	f.CopyData(d.TryData, s.TryData)
	d.cost.Set(s.cost)
}

//UpdateCost recalculates the try cost
func (f *IntFunStub) UpdateCost(t Try) {
	try := t.(*IntTry)
	f.Cost(try.TryData, f.tempCost)
	try.SetCostValue(f.tempCost)
}

//Cmp compares the tries
func (f *IntFunStub) Cmp(x, y Try, mode CmpMode) float64 {
	s := x.(*IntTry)
	t := y.(*IntTry)
	return s.Cmp(t)
}

// ToConstraint uses the previous try pre and the updating hint parameter
// to attempt to produce an update to pre which satisfies
// solution constraints it returns valid = True if succeeds, otherwise pre remains un changed and returns false
func (f *IntFunStub) ToConstraint(pre Try, hint *big.Int) bool {
	p := pre.(*IntTry)
	if f.Constraint(p.TryData, hint) {
		f.SetTry(p, hint)
		return true
	}
	return false
}
