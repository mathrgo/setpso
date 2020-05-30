//Package quadratic includes a sampler for finding algebraic solutions to
//quadratic equation ax^2+bx+c =0.
package quadratic

import (
	"fmt"
	"math/rand"
)

/*Sampler is a sampler that randomly generates test samples of the quadreatic
equation ax^2 + bx + c =0
function. this is to be used together with the FunType type to generate a
function to see how well the SPSO optimisers perform when searching for a
algeraic expression  dag to represent the solution of x to the quadratic equation . */
type Sampler struct {
	maxRange float64
}

//NewSampler creates an instance of the parity sampler for inputSize
//bits.
func NewSampler(inputSize float64) *Sampler {
	p := new(Sampler)
	p.maxRange = 2 * inputSize
	return p
}

//About gives description of sampler
func (p *Sampler) About() string {
	s := fmt.Sprintf("quadratic equation samples for solution x\n")
	s += fmt.Sprintf("with input size %f\n", p.maxRange*0.5)
	return s
}

//Sample generates a sample with input in =[a,b c] and output out = [x].
// where x,a,b are selected randomply from [-inputSize, inputSize).
func (p *Sampler) Sample(in []float64, out []float64, rnd *rand.Rand) {
	//keep +ve ad away from zero
	//a := p.maxRange * (rnd.Float64() + 1.0)
	a := 1.0 // solve sipmlersub proble
	bt := p.maxRange * (rnd.Float64() - 0.5)
	b := 2.0 * a * bt
	// ensure positive squareroot solution
	x := p.maxRange*rnd.Float64() - bt

	c := -(a*x*x + b*x)
	in[0] = b
	in[1] = c
	out[0] = x
}

//InputSize is the number of variables for input.
func (p *Sampler) InputSize() int {
	return 2
}

//OutputSize gives the output size, which in this case is 1.
func (p *Sampler) OutputSize() int {
	return 1
}

//=================================

//ExSampler is an experimental sampler
type ExSampler struct {
	maxRange float64
}

//NewExSampler creates an instance of the parity sampler for inputSize
//bits.
func NewExSampler(maxRange float64) *ExSampler {
	p := new(ExSampler)
	p.maxRange = maxRange
	return p
}

//About gives description of sampler
func (p *ExSampler) About() string {
	s := fmt.Sprintf("experimental equation samples for solution x\n")
	s += fmt.Sprintf("with input size %f\n", p.maxRange*0.5)
	return s
}

//Sample generates a sample.
func (p *ExSampler) Sample(in []float64, out []float64, rnd *rand.Rand) {
	in[0] = (rnd.Float64() - 0.5) * p.maxRange
	in[1] = (rnd.Float64() - 0.5) * p.maxRange
	out[0] = in[0] + in[1]
}

//InputSize is the number of variables for input.
func (p *ExSampler) InputSize() int {
	return 2
}

//OutputSize gives the output size, which in this case is 1.
func (p *ExSampler) OutputSize() int {
	return 1
}
