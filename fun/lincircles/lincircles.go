/*
Package lincircles provides method for packing circles on a straight line using a packing cost
*/
package lincircles

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"text/template"

	"github.com/mathrgo/setpso/psokit"

	"github.com/mathrgo/setpso/fun/futil"
)

//Circle is the type used to represent a packing
type Circle struct {
	x, r float64
}

//CostValue  contains the results of calculating the circle cost
type CostValue struct {
	value       float64
	UsedCircles []Circle
}

//NewCostValue creates a new circles CostValue
func NewCostValue() *CostValue {
	c := new(CostValue)
	c.UsedCircles = make([]Circle, 0, 10)
	return c
}

//Set is used to set c from x
func (c *CostValue) Set(x interface{}) {
	if c1, v := x.(*CostValue); v {
		c.value = c1.value
		c.UsedCircles = c.UsedCircles[:0]
		c.UsedCircles = append(c.UsedCircles, c1.UsedCircles...)
	}
}

//SetValue just sets the value without changing the used circles
func (c *CostValue) SetValue(z float64) {
	c.value = z
}

//AddCircle just adds another circle
func (c *CostValue) AddCircle(x Circle) {
	c.UsedCircles = append(c.UsedCircles, x)

}

//Circles returns the array of used circles
func (c *CostValue) Circles() []Circle {
	return c.UsedCircles
}

//Clear removes the used circles
func (c *CostValue) Clear() {
	c.UsedCircles = c.UsedCircles[:0]
}

//Update adds the mean of x as raw data to calculate an updated stats of the cost value
func (c *CostValue) Update(x interface{}) {
	c.Set(x)
}

//Cmp compares with x
func (c *CostValue) Cmp(x interface{}) int {
	c1 := x.(*CostValue)
	d := c.value - c1.value
	if d > 0 {
		return 1
	} else if d < 0 {
		return -1
	} else {
		return 0
	}
}

//Fbits scales the cost value by taking sign(x.mean)log2(1+|x.mean|)
func (c *CostValue) Fbits() float64 {
	if c.value > 0 {
		return math.Log2(1.0 + c.value)
	}
	return -math.Log2(1 - c.value)
}

// String is human readable value
func (c *CostValue) String() string {
	var s string
	s = fmt.Sprintf(" %f number of used circles =%d\n", c.value,
		len(c.UsedCircles))
	for i := range c.UsedCircles {
		p := c.UsedCircles[i]
		s += fmt.Sprintf("%f\n", p.x)
	}
	return s
}

// Fun is the circles packaging cost function
type Fun struct {
	radius, radius0, radius1 float64
	valueNbits               int
	circles                  []Circle
	d2, d20, d21             float64
	birthBonus               float64
}

//NewCostValue creates a zero cost value
func (f *Fun) NewCostValue() futil.CostValue {

	return NewCostValue()
}

// New generates a circle packing cost function suitable for SPSO.
// 'radius' is the radius of the circles to be packed;
// 'innerFuzz' is reducing factor on 'radius' to get rejection radius
// 'outerFuzz' is increasing factor on 'radius' to get maximum influence radius // 'valueNbits' is the number of bits used to locate a coordinate of a packed
// circle.
// 'birthBonus' is the reduction of cost due to including a circle
func New(radius float64, innerFuzz, outerFuzz float64, valueNbits int, birthBonus float64) *Fun {
	var f Fun
	f.radius = radius
	f.radius0 = radius * (1.0 - innerFuzz)
	f.radius1 = radius * (1.0 + outerFuzz)
	f.valueNbits = valueNbits
	n := int(math.Ceil(1.0 / f.radius))
	f.circles = make([]Circle, n)
	f.d2 = 4.0 * f.radius * f.radius
	f.d20 = 4.0 * f.radius0 * f.radius0
	f.d21 = 4.0 * f.radius1 * f.radius1
	f.birthBonus = birthBonus
	return &f
}

//Len gives the maximum number of circlse in each try
func (f *Fun) Len() int {
	return len(f.circles)
}

// Cost returns the cost of the chosen packing circles
func (f *Fun) Cost(x *big.Int) futil.CostValue {
	f.IDecode(x)
	var cv CostValue
	cv.Clear()
	cost := 0.0
	maxr := 1.0 - f.radius
	maxr2 := maxr * maxr
	dd0 := f.d2 - f.d20
	dd1 := f.d21 - f.d2
	for i := range f.circles {
		c := f.circles[i]
		if maxr2 >= c.x*c.x {

			uc := cv.Circles()
			ok := len(uc) == 0
			overlapped := false
			for i1 := range uc {
				c0 := uc[i1]
				dx := c.x - c0.x
				e := dx * dx
				if e <= f.d21 { // circle is within range of influence
					if e >= f.d2 {
						cost -= (f.d21 - e) / dd1
						ok = true
					} else if e >= f.d20 {
						cost -= (e - f.d20) / dd0
						ok = true
					} else {
						overlapped = true
					}
				}
			}
			if ok && !overlapped {
				cv.AddCircle(c)
				cost -= f.birthBonus
			}

		}

	}
	cv.SetValue(cost)
	return &cv
}

// MaxLen returns the number of elements in the encoded parameter
func (f *Fun) MaxLen() int {
	return 2 * len(f.circles) * f.valueNbits
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
	s = "linear circle packing problem parameters:\n"
	s += fmt.Sprintf("circle  radius = %f  inner radius = %f outer radius %f\n coordinate resolution= %f  birth bonus= %f\n", f.radius, f.radius0, f.radius1, math.Exp2(float64(1-f.valueNbits)), f.birthBonus)

	return s

}

//IDecode  extracts the raw circle positions from z
func (f *Fun) IDecode(z *big.Int) {
	jb := 0
	scale := math.Exp2(1 - float64(f.valueNbits))
	for i := range f.circles {
		var x uint
		f.circles[i].r = f.radius
		x = 0
		je := jb + f.valueNbits
		for j := jb; j < je; j++ {
			x *= 2
			x += z.Bit(j)
		}
		f.circles[i].x = float64(x)*scale - 1.0

		jb += f.valueNbits
	}
}

// Decode requests the function to give a meaningful interpretation of
// z
func (f *Fun) Decode(z *big.Int) string {
	var s string = ""

	return s
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }

//Store stores circles  for plotting
type Store struct {
	dispList  []float64
	skipLen   int
	count     int
	dispScale float64
}

// NewStore  creates a circle store assuming dataLen data points but with
// a skipLen skip interval on this data each snapshot storing atmost nCircles;
// dispSize is the size of animation display in pixels.
func NewStore(dataLen, skipLen, nCircles, dispSize int) *Store {
	var cs Store
	cs.dispList = make([]float64, 0, dataLen*(3*nCircles+1)/skipLen)
	cs.dispScale = float64(dispSize) / 2.0
	return &cs
}

//Clear empties the circle store
func (cs *Store) Clear() {
	cs.count = 0
	cs.dispList = cs.dispList[:0]
}

// Append appends some circlse transformed for display followed by a zero entry
// as an end marker to the list
func (cs *Store) Append(circles []Circle) {
	for i := range circles {
		c := circles[i]
		x := (c.x + 1) * cs.dispScale
		y := cs.dispScale
		r := c.r * cs.dispScale
		cs.dispList = append(cs.dispList, x, y, r)
	}
	cs.dispList = append(cs.dispList, 0.0)
}

//Animate generates a data string for animation
func (cs *Store) Animate() string {
	var s string

	for i := range cs.dispList {
		v := cs.dispList[i]
		if v > 0.0 {
			s += fmt.Sprintf("%f, ", v)
		} else {
			s += fmt.Sprintf(" 0.0,\n")
		}

	}
	return s
}

//Animator works with psokit to generate an HTML file that gives an animation of used circles of global best  during the run.
type Animator struct {
	f        *Fun
	store    *Store
	skipLen  int
	dispSize int
	count    int
}

// NewAnimator creates an animation of the personal best used circles
func NewAnimator(skipLen, dispSize int) *Animator {
	ac := new(Animator)
	ac.skipLen = skipLen
	ac.dispSize = dispSize
	ac.count = 0 // play safe
	return ac
}

//RunInit initialises it for a run
func (ac *Animator) RunInit(man *psokit.ManPso) {
	fmt.Println("Using circle animator")
	if fc, ok := man.F().(*Fun); ok {
		ac.f = fc
		ac.store = NewStore(man.Datalength(), ac.skipLen, ac.f.Len(), ac.dispSize)

	} else {
		panic(fmt.Errorf("cost function is not a circles one"))
	}

}

//DataUpdate  stores used circles of personal best
func (ac *Animator) DataUpdate(man *psokit.ManPso) {
	if ac.count > 0 {
		ac.count--
		return
	}
	ac.count = ac.skipLen
	cost := man.P().GlobalCost()
	c := cost.(*CostValue)
	ac.store.Append(c.Circles())

}

type animData struct {
	DispSize   int
	CircleData string
}

//Result outputs the stored used circles as an animation
func (ac *Animator) Result(man *psokit.ManPso) {
	var data animData
	data.DispSize = ac.dispSize
	data.CircleData = ac.store.Animate()
	tmplText, err := ioutil.ReadFile("animatorTemplate.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl, err := template.New("anim").Parse(string(tmplText))
	if err != nil {
		fmt.Println(err)
		return
	}
	filename := fmt.Sprintf("AnimGlobalBest%d.html", man.RunID())
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Println(err)
	}
	file.Close()

}
