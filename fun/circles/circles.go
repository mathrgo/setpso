/*
Package circles provides method for playing with circles
*/
package circles

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"text/template"

	"github.com/mathrgo/setpso"

	"github.com/mathrgo/setpso/psokit"

	"github.com/mathrgo/setpso/fun/futil"
)

//Circle is the type used to represent a packing
type Circle struct {
	x, y, r float64
}

// Fun is the circles packaging cost function
type Fun struct {
	radius, radius0, radius1 float64
	valueNbits               int
	d2, d20, d21             float64
	birthBonus               float64
	n                        int
}

//Try is the try interface used by setpso
type Try = setpso.Try

//FunTry gives the try structure to use
type FunTry = futil.FloatTry

//TryData is the interface for FunTryData used in package futil
type TryData = futil.TryData

//FunTryData is the decoded data structure for a try
type FunTryData struct {
	UsedCircles []Circle
	circles     []Circle
}

//IDecode decodes z into d
func (f *Fun) IDecode(data TryData, z *big.Int){
	t := data.(*FunTryData)
	jb := 0
	scale := math.Exp2(1 - float64(f.valueNbits))
	for i := range t.circles {
		var x, y uint
		t.circles[i].r = f.radius
		x = 0
		je := jb + f.valueNbits
		for j := jb; j < je; j++ {
			x *= 2
			x += z.Bit(j)
		}
		t.circles[i].x = float64(x)*scale - 1.0

		y = 0
		jb += f.valueNbits
		je = jb + f.valueNbits
		for j := jb; j < je; j++ {
			y *= 2
			y += z.Bit(j)
		}
		t.circles[i].y = float64(y)*scale - 1.0
		jb += f.valueNbits
	}
}

// Decode requests the function to give a meaningful interpretation of
// d.
func (d *FunTryData) Decode() string {
	s := fmt.Sprintf(" number of used circles =%d\n", len(d.UsedCircles))
	for i := range d.UsedCircles {
		c := d.UsedCircles[i]
		s += fmt.Sprintf("circle %d x=%f \ty=%f\n", i, c.x, c.y)
	}
	return s
}

//FloatFunStub gives interface to setpso
type FloatFunStub = futil.FloatFunStub

//New generates a circle packing cost function suitable for SPSO.
// 'radius' is the radius of the circles to be packed;
// 'innerFuzz' is reducing factor on 'radius' to get rejection radius
// 'outerFuzz' is increasing factor on 'radius' to get maximum influence radius // 'value. Nbits' is the number of bits used to locate a coordinate of a packed
// circle.
// 'birthBonus' is the reduction of cost due to including a circle
func New(radius float64, innerFuzz, outerFuzz float64, valueNbits int, birthBonus float64) *FloatFunStub {
	var f Fun
	f.radius = radius
	f.radius0 = radius * (1.0 - innerFuzz)
	f.radius1 = radius * (1.0 + outerFuzz)
	f.valueNbits = valueNbits
	f.n = int(math.Ceil(1.0 / (f.radius * f.radius)))
	f.d2 = 4.0 * f.radius * f.radius
	f.d20 = 4.0 * f.radius0 * f.radius0
	f.d21 = 4.0 * f.radius1 * f.radius1
	f.birthBonus = birthBonus
	return futil.NewFloatFunStub(&f)
}

//CreateData creates a empty structure for decoded try
func (f *Fun) CreateData() TryData {
	t := new(FunTryData)
	t.circles = make([]Circle, f.n)
	t.UsedCircles = make([]Circle, 0, f.n)
	return t
}

//Cost returns the remainder after dividing p in to the prime product
func (f *Fun) Cost(data TryData) (cost float64) {

	d := data.(*FunTryData)
	d.UsedCircles = d.UsedCircles[:0]
	cost = 0.0
	maxr := 1.0 - f.radius
	maxr2 := maxr * maxr
	dd0 := f.d2 - f.d20
	dd1 := f.d21 - f.d2
	for i := range d.circles {
		c := d.circles[i]
		if maxr2 >= c.x*c.x+c.y*c.y {

			uc := d.UsedCircles
			ok := len(uc) == 0
			overlapped := false
			for i1 := range uc {
				c0 := uc[i1]
				dx := c.x - c0.x
				dy := c.y - c0.y
				e := dx*dx + dy*dy
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
				d.UsedCircles = append(d.UsedCircles, c)
				cost -= f.birthBonus
			}

		}

	}
	return
}

//DefaultParam gives a default that satisfies constraints
func (f *Fun) DefaultParam() *big.Int {
	return big.NewInt(0)
}

//CopyData copies src to dest
func (f *Fun) CopyData(dest, src TryData) {
	s := src.(*FunTryData)
	d := dest.(*FunTryData)
	d.circles = d.circles[:0]
	d.circles = append(d.circles, s.circles...)
	d.UsedCircles = d.UsedCircles[:0]
	d.UsedCircles = append(d.UsedCircles, s.UsedCircles...)
}

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return 2 * f.n * f.valueNbits
}

//Constraint attempts to constrain hint possibly using a copy of pre to do this
func (f *Fun) Constraint(pre TryData, hint *big.Int) (valid bool) {
	valid = true
	return
}

// About returns a string description of the contents of Fun
func (f *Fun) About() string {
	var s string
	s = "circle packing problem parameters:\n"
	s += fmt.Sprintf("circle  radius = %f  inner radius = %f outer radius %f\n coordinate resolution= %f  birth bonus= %f\n", f.radius, f.radius0, f.radius1, math.Exp2(float64(1-f.valueNbits)), f.birthBonus)

	return s
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }

//==========================================

//Circles returns the array of used circles
func (d *FunTryData) Circles() []Circle {
	return d.UsedCircles
}

//Len gives the maximum number of circles in each try
func (f *Fun) Len() int {
	return f.n
}

//Store stores circles  for plotting
type Store struct {
	dispList  []float64
	skipLen   int
	count     int
	dispScale float64
}

// NewStore  creates a circle store assuming dataLen data points but with
// a skipLen skip interval on this data each snapshot storing at most nCircles;
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

// Append appends some circles transformed for display followed by a zero entry
// as an end marker to the list
func (cs *Store) Append(circles []Circle) {
	for i := range circles {
		c := circles[i]
		x := (c.x + 1) * cs.dispScale
		y := (c.y + 1) * cs.dispScale
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

//RunInit initializes it for a run
func (ac *Animator) RunInit(man *psokit.ManPso) {
	fmt.Println("Using circle animator")
	fc := man.F().(*FloatFunStub)
	ac.f = fc.Fun().(*Fun)
	ac.store = NewStore(man.Datalength(), ac.skipLen, ac.f.Len(), ac.dispSize)

}

//DataUpdate  stores used circles of personal best
func (ac *Animator) DataUpdate(man *psokit.ManPso) {
	if ac.count > 0 {
		ac.count--
		return
	}
	ac.count = ac.skipLen
	i := man.P().BestParticle()
	data := man.P().LocalBestTry(i).Data().(*FunTryData)
	ac.store.Append(data.Circles())

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
