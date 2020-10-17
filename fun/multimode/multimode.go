//Package multimode is a multimode function with noise
package multimode

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/mathrgo/setpso"
	"github.com/mathrgo/setpso/fun/futil"
)

/*Fun is the multimode cost function to be minimized.
 */
type Fun struct {
	// configuration data
	nMode       int
	nbits       int
	margin      float64
	sigma       float64
	Tc          float64
	SigmaMargin float64
	// internal data
	omega     float64
	slope     float64
	bias      float64
	floatCost float64
	rnd       *rand.Rand
	factor    float64
	bestX     float64
}

//Try is the try interface used by setpso
type Try = setpso.Try

//FunTry gives the try structure to use
type FunTry = futil.SFloatTry

//TryData is the interface for FunTryData used in package futil
type TryData = futil.TryData

//FunTryData is the decoded data structure for a try giving the x coordinate of the function evaluation
type FunTryData struct {
	x float64
}

//IDecode decodes z into da
func (f *Fun) IDecode(data TryData, z *big.Int) {
	t := data.(*FunTryData)
	a := float64(z.Int64())
	t.x = a * f.factor
	//fmt.Printf("x= %f",t.x)
}

//Decode requests the function to give a meaningful interpretation of d
func (d *FunTryData) Decode() string {
	return fmt.Sprintf("x = %f\n", d.x)
}

//SFloatFunStub gives interface to setpso
type SFloatFunStub = futil.SFloatFunStub

/*NewFun creates Fun. Fun is in the range [0,1) to be minimized.
it is of the form
	y=slope*x + bias -sin(w*x) +noise
w is chosen to give nMode local minima slope is chosen to give a difference of margin between local minima and bias is chosen so that the first minimum has the value 0. The noise has zero mean and standard deviation sigma. nbits is the number of bits used to represent x as a parameter to optimise.
margin must be less than 2Pi.
For cost values it uses futil.SFloatCostValue so needs:
Tc -- time constant of statistics in number of cost value updates
*/
func NewFun(nMode, nbits int, margin, sigma float64,
	Tc, SigmaMargin float64, fsd int64) *SFloatFunStub {
	f := new(Fun)
	f.rnd = rand.New(rand.NewSource(fsd))
	f.nMode = nMode
	f.nbits = nbits
	f.margin = margin
	f.sigma = sigma
	f.Tc = Tc
	f.SigmaMargin = SigmaMargin
	f.omega = (float64(f.nMode) + 1.0) * math.Pi
	f.slope = f.omega * f.margin / (2 * math.Pi)
	a := f.slope / f.omega
	f.bias = math.Sqrt(1.0-a*a) - a*math.Acos(a)
	max := 1 << uint(f.nbits)
	f.factor = 1.0 / float64(max)
	f.bestX = math.Acos(a) / f.omega
	return futil.NewSFloatFunStub(f, Tc, SigmaMargin)
}

//CreateData creates a empty structure for decoded try
func (f *Fun) CreateData() TryData {
	return new(FunTryData)

}

//Cost evaluates costvalue
func (f *Fun) Cost(data TryData) (cost float64) {
	d := data.(*FunTryData)
	x := d.x
	cost = f.slope*x + f.bias - math.Sin(f.omega*x)
	cost += f.sigma * f.rnd.NormFloat64()
	return
}

//DefaultParam gives a default that satisfies constraints
func (f *Fun) DefaultParam() *big.Int {
	return big.NewInt(10)
}

//CopyData copies src to dest
func (f *Fun) CopyData(dest, src TryData) {
	s := src.(*FunTryData)
	d := dest.(*FunTryData)
	d.x = s.x
}

//MaxLen  maximum number of bits in the parameter big integer which is the
// maximum number of elements in the subset
func (f *Fun) MaxLen() (maxlen int) {
	return f.nbits
}

//Constraint attempts to constrain hint possibly using a copy of pre to do this
func (f *Fun) Constraint(pre TryData, hint *big.Int) (valid bool) {
	valid = true
	return
}

// About gives a description of the cost function
func (f *Fun) About() (s string) {
	s = fmt.Sprintf("multimode with noise test function\n")
	s += fmt.Sprintf("number of minima = %d  resolution in bits = %d\n", f.nMode, f.nbits)
	s += fmt.Sprintf("local minima margin = %f ", f.margin)
	s += fmt.Sprintf("noise sigma = %f\n", f.sigma)
	s += fmt.Sprintf("stats time constant = %f ", f.Tc)
	s += fmt.Sprintf(" comparison margin in sigmas = %f\n", f.SigmaMargin)
	s += fmt.Sprintf("best value for x = %f\n", f.bestX)
	return s
}

//Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }
