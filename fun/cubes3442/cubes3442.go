/*
Package cubes3442 is a function for finding the sum of 3  cubes that equals 42.
it uses the methods described in https://github.com/mathrgo/setpso/blob/master/doc/sumcubesof42.pdf .
*/
package cubes3442

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso/fun/futil"
)

//Fun is the data structure used
type Fun struct {
	bitLen                      int
	x0, x1, x2                  *big.Int
	sp                          *futil.Splitter
	parts                       []*big.Int
	zero, one, two, three, four *big.Int
	six, seven                  *big.Int
	fortyTwo                    *big.Int
	cost                        futil.CostValue
}

//NewCostValue creates a zero cost value representing a
// big integer.
func (f *Fun) NewCostValue() futil.CostValue {
	return futil.NewIntCostValue()
}

/*New creates an instance of the function that uses bitLen bits for the positive
candidate integers j0,j1,j2 in the encoding.
*/
func New(bitLen int) *Fun {
	var f Fun
	f.bitLen = bitLen
	f.x0 = big.NewInt(0)
	f.x1 = big.NewInt(0)
	f.x2 = big.NewInt(0)
	f.sp = futil.NewSplitter(bitLen, bitLen, bitLen, 5)
	f.zero = big.NewInt(0)
	f.one = big.NewInt(1)
	f.two = big.NewInt(2)
	f.three = big.NewInt(3)
	f.four = big.NewInt(4)
	f.six = big.NewInt(6)
	f.seven = big.NewInt(7)
	f.fortyTwo = big.NewInt(42)
	f.cost = f.NewCostValue()
	f.parts = make([]*big.Int, 4)
	for i := range f.parts {
		f.parts[i] = big.NewInt(0)
	}
	return &f
}

// BitLen returns the number of bits used to encode each integer part.
func (f *Fun) BitLen() int {
	return f.bitLen
}

// this just extracts the x0, x1,x2 possibly prior to constraint satisfaction.
func (f *Fun) evalXs(x *big.Int) {
	c3index := 0
	f.parts = f.sp.Split(x, f.parts)
	// evaluate x0
	f.x0.Mul(f.six, f.parts[0])
	if f.parts[3].Bit(0) == 1 {
		f.x0.Neg(f.x0)
	}
	if f.parts[3].Bit(3) == 1 {
		c3index++
		f.x0.Add(f.x0, f.two)
	} else {
		f.x0.Sub(f.x0, f.one)
	}
	f.x0.Mul(f.x0, f.seven)
	f.evalX12(c3index)
}

// calculates sum of cubes given evalXs values
func (f *Fun) ans() *big.Int {
	var cube big.Int
	cost := big.NewInt(0)
	// add cube of x0
	cube.Mul(cube.Mul(f.x0, f.x0), f.x0)
	cost.Add(cost, &cube)
	// add cube of x1
	cube.Mul(cube.Mul(f.x1, f.x1), f.x1)
	cost.Add(cost, &cube)
	// add cube of x2
	cube.Mul(cube.Mul(f.x2, f.x2), f.x2)
	cost.Add(cost, &cube)
	return cost
}

// evaluates x1,x2 assuming f.parts have been extracted
func (f *Fun) evalX12(c3index int) {
	// evaluate x1
	f.x1.Mul(f.six, f.parts[1])
	if f.parts[3].Bit(1) == 1 {
		f.x1.Neg(f.x1)
	}
	if f.parts[3].Bit(4) == 1 {
		c3index += 2
		f.x1.Add(f.x1, f.two)
	} else {
		f.x1.Sub(f.x1, f.one)
	}
	// evaluate x2
	f.x2.Mul(f.six, f.parts[2])
	if f.parts[3].Bit(2) == 1 {
		f.x2.Neg(f.x2)
	}
	switch c3index {
	case 0, 3:
		f.x2.Add(f.x2, f.two)
	case 1, 2:
		f.x2.Sub(f.x2, f.one)
	}
}

/*
Cost returns the absolute value of the difference between decoded
sum of cubes and 42. It assumes x already meets constraints.
*/
func (f *Fun) Cost(x *big.Int) futil.CostValue {
	f.evalXs(x)
	cost1 := f.ans()
	// calculate error
	f.cost.Set(cost1.Abs(cost1.Sub(cost1, f.fortyTwo)))
	return f.cost
}

// MaxLen returns the maximum number of bits to use for parameter x
func (f *Fun) MaxLen() int {
	return f.sp.MaxBits()
}

// ToConstraint uses the previous parameter pre and the updating hint parameter
// to attempt to produce an update to hint which satisfies solution constraints
// and returns valid = True if succeeds
func (f *Fun) ToConstraint(pre, hint *big.Int) (valid bool) {
	valid = true
	f.parts = f.sp.Split(hint, f.parts)
	c3index := 0
	if f.parts[3].Bit(3) == 1 {
		c3index++
	}
	f.evalX12(c3index)
	var m big.Int
	// correct for mod 7 constraint on x1
	// do not bother to check for sign change in new x1
	m.Mod(f.x1, f.seven)
	mint := m.Int64()
	var d big.Int
	if f.parts[3].Bit(1) == 0 {
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
	f.parts[1].Add(f.parts[1], &d)

	// correct for mod 7 constraint on x2
	// do not bother to check for sign change in new x1
	m.Mod(f.x2, f.seven)
	mint = m.Int64()
	if f.parts[3].Bit(2) == 0 {
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
	f.parts[2].Add(f.parts[2], &d)
	if f.sp.Join(f.parts, hint) != nil {
		valid = false
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

// Decode requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (f *Fun) Decode(z *big.Int) (s string) {
	f.evalXs(z)
	ans := f.ans()
	s = fmt.Sprintf("x0: %v\n", f.x0)
	s += fmt.Sprintf("x1: %v\n", f.x1)
	s += fmt.Sprintf("x2: %v\n", f.x2)
	mod := big.NewInt(0)
	div := big.NewInt(0)
	div.DivMod(ans, f.fortyTwo, mod)
	s += fmt.Sprintf("ans/42: %v\n", div)
	s += fmt.Sprintf("ans mod 42: %v\n", mod)
	s += fmt.Sprintf("part0: %x \n", f.parts[0])
	s += fmt.Sprintf("part1: %x \n", f.parts[1])
	s += fmt.Sprintf("part2: %x \n", f.parts[2])
	s += fmt.Sprintf("flags %b \n", f.parts[3])

	return
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }
