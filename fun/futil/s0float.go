package futil
import(
	"fmt"
	"math"

)


//===================================================================//

/*S0FloatCostValue is an old data type that is no longer used to store Float   cost values based on simple statistics. 
Internally it maintains an mean as the cost plus the
  variance of the cost, which is used to determine if one cost is bigger than
  another. This is communicated using the Cmp function. As well as this it uses an updating weight that incorporates a forgetting process that allows for
  gradual change to the cost function itself. */
  type S0FloatCostValue struct {
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
}
// Copy takes a copy of c1
func (c *S0FloatCostValue)Copy(c1 *S0FloatCostValue){
	// play safe by using explicit copy
	c.mean =c1.mean
	c.variance=c1.variance
	c.bMem = c1.bMem
	c.lambda =c1.lambda
	c.minLambda = c1.minLambda
	c.delta = c1.delta
	c.isum = c1.isum
	c.Tc = c1.Tc
}

// NewS0FloatCostValue returns a pointer to S0FloatCostValue type initialized
// with a  time constant to stats change of Tc.
func NewS0FloatCostValue(Tc float64) *S0FloatCostValue {

	c := new(S0FloatCostValue)
	c.Tc = Tc
	c.minLambda = 1.0 / Tc

	return c
}

//String gives human readable description
func (c *S0FloatCostValue) String() string {
	return fmt.Sprintf("mean=%f variance=%f  TC=%f  \n ",
		c.mean, c.variance, c.Tc)
}

//Set is used to set c from x
func (c *S0FloatCostValue) Set(x float64) {
	c.mean = x
	c.variance = 1.0 // play safe and remove from singularity at zero variance
	c.bMem = 1.0
	c.lambda = 1.0
	c.delta = 1.0
	c.isum = x * x
}

//Update adds the mean of x as raw data to calculate an updated stats of the cost value
func (c *S0FloatCostValue) Update(x float64) {
	if c.variance <= math.SmallestNonzeroFloat64 {
		c.Set(x)
		return
	}
	c.lambda = c.lambda / (c.lambda + c.bMem)
	if c.lambda < c.minLambda {
		c.lambda = c.minLambda
		c.bMem = 1.0 - c.minLambda
	}
	c.delta = c.bMem*c.bMem*c.delta + 1.0
	l1 := 1.0 - c.lambda
	l0 := c.lambda
	c.mean = l1*c.mean + l0*x
	c.isum = l1*c.isum + l0*x*x
	if dl := c.delta * l0 * l0; dl < 1.0 {
		c.variance = (c.isum - c.mean*c.mean) * dl / (1.0 - dl)
	}

}

//Cmp compares with x
func (c *S0FloatCostValue) Cmp(c1 *S0FloatCostValue) float64 {
	d := c.mean - c1.mean
	d2:=d*d/(c.variance+c1.variance)
	if d>0 {
		return d2
	}
	return -d2	
}

// Fbits scales the cost value by taking sign(x.mean)log2(1+|x.mean|)
func (c *S0FloatCostValue) Fbits() float64 {
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
