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

	"github.com/mathrgo/setpso"
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
	//NBit is the number of bits used to give the element values
	NBit int
	//Seed used for generating the subset sum problems
	Seed int64
}

//Try is the try interface used by setpso
type Try = setpso.Try

//FunTry gives the try structure to use 
type FunTry=futil.IntTry
//TryData is the interface for FunTryData used in package futil
type TryData= futil.TryData

//FunTryData is the decoded data structure for a try
type FunTryData struct {
	subset *big.Int
}

//IDecode decodes z into data
func (f *Fun) IDecode(data TryData, z *big.Int) {
	data.(*FunTryData).subset.Set(z)
	//fmt.Printf("subset= %v\n",d.subset)
}

// Decode requests the try to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (d *FunTryData) Decode() string {
	return d.subset.Text(2)
}
//IntFunStub gives interface to setpso
type IntFunStub = futil.IntFunStub

// New generates a subset sum problem using nElement set elements and
// nBit bits to represent the element values using the random number
// generator seed sd , where all element values are taken to be non negative
func New(nElement int, nBit int, sd int64) *IntFunStub {
	var f Fun
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

	return futil.NewIntFunStub(&f)
}

//CreateData creates a empty structure for decoded try
func (f *Fun)CreateData()TryData{
t:=new(FunTryData)
t.subset=new(big.Int)
return t
}

// Cost calculates the cost of t as  the absolute value of the difference
// between the subset sum and the target value for the subset try t
func (f *Fun) Cost(data TryData, cost *big.Int) {
	x := data.(*FunTryData).subset
	cost.SetInt64(0)
	for i := range f.ElementValues {
		if x.Bit(i) == 1 {
			cost.Add(cost, f.ElementValues[i])
		}
	}
	cost.Abs(cost.Sub(cost, f.Target))
}

//DefaultParam gives a default that satisfies constraints
func (f *Fun)DefaultParam() *big.Int{
	return new(big.Int)
}

//CopyData copies src to dest
func (f *Fun)CopyData(dest,src TryData){
	s:=src.(*FunTryData)
	d:=dest.(*FunTryData)
d.subset.Set(s.subset)
}

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return len(f.ElementValues)
}
//Constraint attempts to constrain hint possibly using a copy of pre to do this
func (f *Fun) Constraint(pre TryData, hint *big.Int) (valid bool) {
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

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }
