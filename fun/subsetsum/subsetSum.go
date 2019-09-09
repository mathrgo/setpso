// Package subsetsum contains the object Fun and its generator for the
// the cost function of the combinatorial subset sum problem
// where a big integer is used to represent subsets as well as
// integer values thus potentially being able to represent
// combinatorial hard problems
package subsetsum

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/mathrgo/setpso/fun/futil"
)

// Fun is the subset sum problem cost function
type Fun struct {
	// Target is the subset sum to look for
	Target *big.Int
	// targetS is a solution possibly among many
	targetS *big.Int
	// ElementValues is the array of values of the corresponding elements
	// in the set where the ith element corresponds to the ith bit of the
	// big integer  representing the subset
	ElementValues []*big.Int
	// scratch pad for calculating sums of element values
	sum *big.Int
	//NBit is the number of bits used to give the element values
	NBit int
	//Seed used for generating the subset sum problems
	Seed int64
	//Store of CostValue
	cost futil.CostValue
}

//NewCostValue creates a zero cost value representing a
// big integer.
func (f *Fun) NewCostValue() futil.CostValue {
	return futil.NewIntCostValue()
}

// New generates a subset sum problem using n Element set elements and
// nBit bits to represent the element values using the random number
// generator seed sd , where all element values are taken to be non negative
func New(nElement int, nBit int, sd int64) *Fun {
	var f Fun
	f.sum = big.NewInt(0)
	f.Seed = sd
	f.NBit = nBit
	f.ElementValues = make([]*big.Int, nElement)
	rnd := rand.New(rand.NewSource(sd))
	maxVal := big.NewInt(0)
	maxVal.SetBit(maxVal, nBit, 1)
	maxVal.Sub(maxVal, big.NewInt(1))
	for i := range f.ElementValues {
		f.ElementValues[i] = big.NewInt(0)
		f.ElementValues[i].Rand(rnd, maxVal)
	}
	// choose number of elements to find for  target subset
	n := rnd.Intn(nElement) + 1
	f.Target = big.NewInt(0)
	f.targetS = big.NewInt(0)
	for i := 0; i < n; i++ {
		j := rnd.Intn(nElement)
		if f.targetS.Bit(j) == 0 {
			f.targetS.SetBit(f.targetS, j, 1)
			f.Target.Add(f.Target, f.ElementValues[j])
		}
	}
	f.cost = f.NewCostValue()
	return &f
}

// Cost returns the absolute value of the difference between the subset sum
// and the target value for the sum for the subset x
func (f *Fun) Cost(x *big.Int) futil.CostValue {
	f.sum.SetInt64(0)
	for i := range f.ElementValues {
		if x.Bit(i) == 1 {
			f.sum.Add(f.sum, f.ElementValues[i])
		}
	}
	f.cost.Set(f.sum.Abs(f.sum.Sub(f.sum, f.Target)))
	return f.cost
}

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return len(f.ElementValues)
}

// ToConstraint uses the previous parameter pre and the updating hint parameter
// to attempt to produce an update to hint which satisfies solution constraints
// and returns valid = True if succeeds
func (f *Fun) ToConstraint(pre, hint *big.Int) (valid bool) {
	valid = true
	return
}

// About returns a string description of the contents of Fun
func (f *Fun) About() string {
	var s string
	s = "subset value problem parameters:\n"
	s += fmt.Sprintf("nElements= %d NBit = %d Seed= %v\n",
		f.MaxLen(), f.NBit, f.Seed)
	s += fmt.Sprintf("Target value: %v\n", f.Target)
	s += "subset solution:\n"
	s += fmt.Sprintf("%s\n", f.targetS.Text(2))
	s += "Values:\n"
	for i := range f.ElementValues {
		s += fmt.Sprintf("%d \t %v\n", i, f.ElementValues[i])

	}

	return s
}

// Decode requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (f *Fun) Decode(z *big.Int) (s string) {
	return z.Text(2)
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }
