package dag

import (
	"fmt"
	"math/big"
	"math/rand"
)

/*ParitySampler is a sampler that randomly generates test samples of the parity
function. this is to be used together with the FunBool type to generate a
function to see how well the SPSO optimisers perform when searching for a
boolean dag to represent the the parity checker. */
type ParitySampler struct {
	maxIn     *big.Int
	inputSize int
	count     int64
}

//NewParitySampler creates an instance of the parity sampler for inputSize
//bits.
func NewParitySampler(inputSize int) *ParitySampler {
	p := new(ParitySampler)
	p.count = 0
	p.maxIn = big.NewInt(1)
	p.maxIn.Lsh(p.maxIn, uint(inputSize))
	//fmt.Printf("maxIn: %b\n", p.maxIn)
	p.inputSize = inputSize
	return p
}

//About gives description of sampler
func (p *ParitySampler) About() string {
	s := fmt.Sprintf("parity samples for %d inputs\n", p.inputSize)
	return s
}

//Sample generates a sample with input x and output y.
func (p *ParitySampler) Sample(x, y *big.Int, rnd *rand.Rand) {

	x.SetInt64(p.count)
	//fmt.Printf("x= %b ", x)
	parity := true
	for i := 0; i < x.BitLen(); i++ {
		if x.Bit(i) == 1 {
			parity = !parity
		}
	}
	if parity {
		y.SetInt64(0)
	} else {
		y.SetInt64(1)
	}
	p.count++
	if p.count >= p.maxIn.Int64() {
		p.count = 0
	}
}

//InputSize is the number of bits to parity check.
func (p *ParitySampler) InputSize() int {
	return p.inputSize
}

//OutputSize gives the output size in bits, which in this case is 1.
func (p *ParitySampler) OutputSize() int {
	return 1
}
