/*
Package cubes3442 is a function for finding the sum of 3  cubes that equals 42.
it uses the methods described in https://github.com/mathrgo/setpso/blob/master/doc/sumcubesof42.pdf .
*/
package cubes3442

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso"

	"github.com/mathrgo/setpso/fun/futil"
)

//Fun is the data structure used
type Fun struct {
	bitLen int

	sp *futil.Splitter

	zero, one, two, three, four *big.Int
	six, seven                  *big.Int
	fortyTwo                    *big.Int
	temp                        TryData
	count                       int
}

//Try is the try interface used by setpso
type Try = setpso.Try

//FunTry gives the try structure to use
type FunTry = futil.IntTry

//TryData is the interface for FunTryData used in package futil
type TryData = futil.TryData

//FunTryData is the decoded data structure for a try
type FunTryData struct {
	x0, x1, x2 *big.Int
	parts      []*big.Int
}

//IDecode decodes z into data
func (f *Fun) IDecode(data TryData, z *big.Int) {
	// this just extracts the x0, x1,x2 possibly prior to constraint satisfaction.
	d := data.(*FunTryData)
	c3index := 0
	d.parts = f.sp.Split(z, d.parts)
	// evaluate x0
	d.x0.Mul(f.six, d.parts[0])
	if d.parts[3].Bit(0) == 1 {
		d.x0.Neg(d.x0)
	}
	if d.parts[3].Bit(3) == 1 {
		c3index++
		d.x0.Add(d.x0, f.two)
	} else {
		d.x0.Sub(d.x0, f.one)
	}
	d.x0.Mul(d.x0, f.seven)
	d.evalX12(c3index, f)

}

// Decode requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (f *FunTryData) Decode() (s string) {
	var sum big.Int
	fortyTwo := big.NewInt(42)
	f.ans(&sum)
	s = fmt.Sprintf("x0: %v\n", f.x0)
	s += fmt.Sprintf("x1: %v\n", f.x1)
	s += fmt.Sprintf("x2: %v\n", f.x2)
	mod := big.NewInt(0)
	div := big.NewInt(0)
	div.DivMod(&sum, fortyTwo, mod)
	s += fmt.Sprintf("ans/42: %v\n", div)
	s += fmt.Sprintf("ans mod 42: %v\n", mod)
	s += fmt.Sprintf("part0: %x \n", f.parts[0])
	s += fmt.Sprintf("part1: %x \n", f.parts[1])
	s += fmt.Sprintf("part2: %x \n", f.parts[2])
	s += fmt.Sprintf("flags %b \n", f.parts[3])

	return
}

//IntFunStub gives interface to setpso
type IntFunStub = futil.IntFunStub

/*New creates an instance of the function that uses bitLen bits for the positive
candidate integers j0,j1,j2 in the encoding.
*/
func New(bitLen int) *IntFunStub {
	var f Fun
	f.bitLen = bitLen

	f.sp = futil.NewSplitter(bitLen, bitLen, bitLen, 5)
	f.zero = big.NewInt(0)
	f.one = big.NewInt(1)
	f.two = big.NewInt(2)
	f.three = big.NewInt(3)
	f.four = big.NewInt(4)
	f.six = big.NewInt(6)
	f.seven = big.NewInt(7)
	f.fortyTwo = big.NewInt(42)
	f.temp = f.CreateData()

	return futil.NewIntFunStub(&f)
}

//CreateData creates a empty structure for decoded try
func (f *Fun) CreateData() TryData {
	t := new(FunTryData)
	t.x0 = big.NewInt(0)
	t.x1 = big.NewInt(0)
	t.x2 = big.NewInt(0)
	t.parts = make([]*big.Int, 4)
	for i := range t.parts {
		t.parts[i] = big.NewInt(0)
	}
	return t
}

/*
Cost returns the absolute value of the difference between decoded
sum of cubes and 42. It assumes x already meets constraints.
*/
func (f *Fun) Cost(data TryData, cost *big.Int) {
	d := data.(*FunTryData)
	d.ans(cost)
	// calculate error
	cost.Abs(cost.Sub(cost, f.fortyTwo))
}

//DefaultParam gives a default that satisfies constraints
func (f *Fun) DefaultParam() *big.Int {

	return big.NewInt(23456)
}

//CopyData copies src to dest
func (f *Fun) CopyData(dest, src TryData) {
	s := src.(*FunTryData)
	d := dest.(*FunTryData)
	d.x0.Set(s.x0)
	d.x1.Set(s.x1)
	d.x2.Set(s.x2)
	for i := range s.parts {
		d.parts[i].Set(s.parts[i])
	}
}

// MaxLen returns the maximum number of bits to use for parameter x
func (f *Fun) MaxLen() int {
	return f.sp.MaxBits()
}

// Constraint uses the previous parameter pre and the updating hint parameter
// to attempt to produce an update to hint which satisfies solution constraints
// and returns valid = True if succeeds
func (f *Fun) Constraint(pre TryData, hint *big.Int) (valid bool) {
	valid = true
	temp := f.temp.(*FunTryData)
	temp.parts = f.sp.Split(hint, temp.parts)
	c3index := 0
	if temp.parts[3].Bit(3) == 1 {
		c3index++
	}
	temp.evalX12(c3index, f)
	var m big.Int
	// correct for mod 7 constraint on x1
	// do not bother to check for sign change in new x1
	m.Mod(temp.x1, f.seven)
	mint := m.Int64()
	var d big.Int
	if temp.parts[3].Bit(1) == 0 {
		switch mint {
		case 1, 2, 4:
			d.Set(f.zero)
		case 3, 5:
			d.Set(f.one)
		case 6:
			d.Set(f.two)
		case 0:
			d.Set(f.three)
		}
	} else {
		switch mint {
		case 1, 2, 4:
			d.Set(f.zero)
		case 0, 3:
			d.Set(f.one)
		case 6:
			d.Set(f.two)
		case 5:
			d.Set(f.three)
		}
	}
	temp.parts[1].Add(temp.parts[1], &d)

	// correct for mod 7 constraint on x2
	// do not bother to check for sign change in new x1
	m.Mod(temp.x2, f.seven)
	mint = m.Int64()
	if temp.parts[3].Bit(2) == 0 {
		switch mint {
		case 3, 5, 6:
			d.Set(f.zero)
		case 0, 4:
			d.Set(f.one)
		case 1:
			d.Set(f.two)
		case 2:
			d.Set(f.three)
		}
	} else {
		switch mint {
		case 3, 5, 6:
			d.Set(f.zero)
		case 2, 4:
			d.Set(f.one)
		case 1:
			d.Set(f.two)
		case 0:
			d.Set(f.three)
		}
	}
	temp.parts[2].Add(temp.parts[2], &d)
	if f.sp.Join(temp.parts, hint) != nil {
		valid = false
		fmt.Printf("invalid\n")
	}
	return
}

// About returns a string description of the contents of Fun
func (f *Fun) About() string {
	var s string
	s = "sum of 3 cubes equal to 42 parameters:\n"
	s += fmt.Sprintf("bitwidth for each cubes encoding: %d\n", f.BitLen())
	return s
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }

// BitLen returns the number of bits used to encode each integer part.
func (f *Fun) BitLen() int {
	return f.bitLen
}

// calculates sum of cubes
func (f *FunTryData) ans(sum *big.Int) {
	var cube big.Int
	sum.SetInt64(0)
	// add cube of x0
	cube.Mul(cube.Mul(f.x0, f.x0), f.x0)
	sum.Add(sum, &cube)
	// add cube of x1
	cube.Mul(cube.Mul(f.x1, f.x1), f.x1)
	sum.Add(sum, &cube)
	// add cube of x2
	cube.Mul(cube.Mul(f.x2, f.x2), f.x2)
	sum.Add(sum, &cube)

}

// evaluates x1,x2 assuming f.parts have been extracted
func (f *FunTryData) evalX12(c3index int, fn *Fun) {
	// evaluate x1
	f.x1.Mul(fn.six, f.parts[1])
	if f.parts[3].Bit(1) == 1 {
		f.x1.Neg(f.x1)
	}
	if f.parts[3].Bit(4) == 1 {
		c3index += 2
		f.x1.Add(f.x1, fn.two)
	} else {
		f.x1.Sub(f.x1, fn.one)
	}
	// evaluate x2
	f.x2.Mul(fn.six, f.parts[2])
	if f.parts[3].Bit(2) == 1 {
		f.x2.Neg(f.x2)
	}
	switch c3index {
	case 0, 3:
		f.x2.Add(f.x2, fn.two)
	case 1, 2:
		f.x2.Sub(f.x2, fn.one)
	}
}
