//Package multimode is a multimode function with noise
 package multimode
import (
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/mathrgo/setpso/fun/futil"
)
/*Fun is the multimode cost function to be minimised.
 */
 type Fun struct {
	// configuration data
	nmode          int
	nbits          int
	margin         float64
	sigma          float64
	Tc, sigmaThres float64
	// intenal data
	omega     float64
	slope     float64
	bias      float64
	floatCost float64
	cost      *futil.SFloatCostValue
	rnd       *rand.Rand
	factor    float64
	bestX     float64
}

/*NewFun creates Fun. Fun is in the range [0,1) to be minimised.
it is of the form
	y=slope*x + bias -sin(w*x) +noise
w is chosen to give nmode local minima slope is chosen to give a difference of margin between local minima and bias is chosen so that the first minimum has the value 0. The noise has zero mean and standard deviation sigma. nbits is the number of bits used to represent x as a parameter to optimise.
margin must be less than 2Pi.
For cost values it uses futil.SFloatCostValue so needs:
Tc -- time constant of statistics in number of cost value updates
sigmaThres -- theshold of comparison certainty in sigmas  of mean measurement.
*/
func NewFun(nmode, nbits int, margin, sigma float64,
	Tc, sigmaThres float64, fsd int64) *Fun {
	f := new(Fun)
	f.rnd = rand.New(rand.NewSource(fsd))
	f.nmode = nmode
	f.nbits = nbits
	f.margin = margin
	f.sigma = sigma
	f.Tc = Tc
	f.sigmaThres = sigmaThres
	f.omega = (float64(f.nmode) + 1.0) * math.Pi
	f.slope = f.omega * f.margin / (2 * math.Pi)
	a := f.slope / f.omega
	f.bias = math.Sqrt(1.0-a*a) - a*math.Acos(a)
	f.cost = futil.NewSFloatCostValue(f.Tc, f.sigmaThres)
	maxx := 1 << uint(f.nbits)
	f.factor = 1.0 / float64(maxx)
	f.bestX = math.Acos(a) / f.omega
	return f
}

//IDecode converts the parameter x to a floating point value in the range [0,1)
func (f *Fun) IDecode(x *big.Int) float64 {
	a := float64(x.Int64())
	b := a * f.factor
	//fmt.Printf("x= %f",b)
	return b

}

//Cost evaluate costvalue
func (f *Fun) Cost(z *big.Int) futil.CostValue {
	x := f.IDecode(z)
	c := f.slope*x + f.bias - math.Sin(f.omega*x)
	c += f.sigma * f.rnd.NormFloat64()
	f.cost.Set(c)
	return f.cost
}

//MaxLen  maximum number of bits in the parameter big integer which is the
// maximum number of elements in the subset
func (f *Fun) MaxLen() (maxlen int) {
	return f.nbits
}

// About gives a description of the cost function
func (f *Fun) About() (s string) {
	s = fmt.Sprintf("multimode with noise function to test util.SFloatCostValue\n")
	s += fmt.Sprintf("number of minima = %d  resolution in bits = %d\n", f.nmode, f.nbits)
	s += fmt.Sprintf("local minima margin = %f ", f.margin)
	s += fmt.Sprintf("noise sigma = %f\n", f.sigma)
	s += fmt.Sprintf("stats time constant = %f ", f.Tc)
	s += fmt.Sprintf(" difference detecting shreshold in measurement sigmas = %f\n", f.sigmaThres)
	s += fmt.Sprintf("best value for x = %f", f.bestX)
	return s
}

//ToConstraint attempts to  give a constraint satisfying hint that matches the hint;
// pre is the previous constraint satisfying version to hint, which
// should not be changed
func (f *Fun) ToConstraint(pre, hint *big.Int) bool { return true }

//Decode requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (f *Fun) Decode(z *big.Int) (s string) {
	fn := f.IDecode(z)
	s = fmt.Sprintf("x = %f\n", fn)
	return s
}

//Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }

//NewCostValue creates zero cost value
func (f *Fun) NewCostValue() futil.CostValue {
	return futil.NewSFloatCostValue(f.Tc, f.sigmaThres)
}
