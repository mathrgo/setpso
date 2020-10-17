/*
Package futil provides some usefull utilities for cost functions
*/
package futil

import (
	"fmt"
	"math/big"
	"math/bits"
)

//W is  assumed size of word used  in big int array
const W = bits.UintSize

/*
Splitter is used to split a positive big int up into an array of big ints parts formed
from the big int. For Computational simplicity the original integer absolute
value is split up into sub slices of big.Word (assumed to be of uint); the
sign of the original big int is not used.
*/
type Splitter struct {
	offset      []int
	wordWidth   []int
	zeroingMask []big.Word
	// maximum number of bits  used including  padding to achieve word alignment.
	maxBits int
	// number of bits for each part.
	bits []int
	// required length in words.
	len int
	// array of zeros to append when needed
	zeros []big.Word
}

/* String gives a readable description of internal state mainly for diagnosis.
 */
func (sp *Splitter) String() string {
	var s string
	s += fmt.Sprintf("offsets:\n")
	for i := range sp.offset {
		s += fmt.Sprintf(" %d", sp.offset[i])
	}
	s += fmt.Sprintf("\nWidth in words:\n")
	for i := range sp.wordWidth {
		s += fmt.Sprintf(" %d", sp.wordWidth[i])

	}
	s += fmt.Sprintf("\n Zero mask:\n")
	for i := range sp.zeroingMask {
		s += fmt.Sprintf(" %x\n", sp.zeroingMask[i])
	}

	s += fmt.Sprintf("Maximum number of bits: %d\n", sp.maxBits)
	s += fmt.Sprintf("Part size in Bits:\n")
	for i := range sp.bits {
		s += fmt.Sprintf(" %d", sp.bits[i])
	}
	s += fmt.Sprintf("\nNumber of words: %d", sp.len)

	return s
}

/*
NewSplitter creates a new Splitter. Bits is an array of number of used bits for
each split element. T
*/
func NewSplitter(bits ...int) *Splitter {
	var s Splitter
	n := len(bits)
	s.bits = make([]int, n)
	s.offset = make([]int, n)
	s.wordWidth = make([]int, n)
	s.zeroingMask = make([]big.Word, n)

	origin := 0
	for i := range bits {
		s.offset[i] = origin
		x := bits[i]
		s.bits[i] = x
		k := (x / W)
		r := uint(x % W)
		if r > 0 {
			k++
			s.zeroingMask[i] = big.Word(1<<r - 1)
		} else {
			s.zeroingMask[i] = ^big.Word(0)
		}
		s.wordWidth[i] = k
		origin += k
	}
	s.len = origin
	s.maxBits = W * origin
	s.zeros = make([]big.Word, s.len)
	return &s
}

// MaxBits returns the maximum bits used by the splitter.
func (sp *Splitter) MaxBits() int {
	return sp.maxBits
}

/*Split takes a copy of the absolute value of x and splits it up into positive
big int parts in place ensuring the parts are word aligned. While doing this it
modifies the copy to match the splitting so that each part is a sub slice of
the modified copy. To do this unused padding bits are set to zero. .The parts
array is overwritten and is mapped into this copy. */
func (sp *Splitter) Split(x *big.Int, parts []*big.Int) []*big.Int {
	// do a manipulation on the local x to ensure its absolute value has enough
	// words for a direct read using Bits()
	y := big.NewInt(0)
	y.Set(x)
	words := y.Bits()
	if len(parts) != sp.len {
		parts = make([]*big.Int, sp.len)
		for i := range parts {
			parts[i] = big.NewInt(0)
		}
	}
	l := len(words)
	if l < sp.len {
		words = append(words, sp.zeros[l:sp.len]...)
	}
	for i := range sp.bits {
		begin := sp.offset[i]
		end := begin + sp.wordWidth[i]
		words[end-1] &= sp.zeroingMask[i]
		parts[i].SetBits(words[begin:end])
	}
	return parts
}

/*Join creates a positive big int from a list of word aligned parts using
splitter bit sizes and writes the result into x. it returns an error if the
parts are too large to fit or number of parts is not compatible. Also when
there is an error x remains unchanged. */
func (sp *Splitter) Join(parts []*big.Int, x *big.Int) (err error) {
	if len(parts) != sp.len {
		return fmt.Errorf("joining parts should have %d members\n ", sp.len)
	}
	// check for part fitting before transferring to x
	for i := range sp.bits {
		if sp.bits[i] < parts[i].BitLen() {
			return fmt.Errorf("part %d is too big to fit", i)
		}
	}
	//empty x to receive parts
	x.SetInt64(0)
	words := x.Bits()
	words = append(words, sp.zeros...)
	for i := range parts {
		w := parts[i].Bits()
		begin := sp.offset[i]
		end := begin + len(w)
		copy(words[begin:end], w)
	}
	x.SetBits(words)
	return nil
}

//==============================================

//CmpMode indicates modes for Cmp.
type CmpMode int

const (
	//CostMode :Cmp compares costs
	CostMode CmpMode = iota
	//TriesMode : Cmp compares tries with best try
	TriesMode
)

/*
Try is an interface  that is used by  a structure to store a big integer parameter, internal decoding of the parameter, and current evaluated cost of a particular try at finding a good solution to the optimization.The interface is usually used to store  data that depends on the parameter and is mainly manipulated by the cost function.
*/
type Try interface {
	// returns floating point representation of
	// of number of bits needed to represent cost as if an integer.
	Fbits() float64
	// subset parameter used by Try
	Parameter() *big.Int
	// this  gives a human readable interpretation of Try based on the internal decoding of the Parameter
	Decode() string
	//this gives human readable cost details such as variance ....
	Cost() string
	// decoded data part
	Data() TryData
}

//TryData is used to store internal decoded data of a try.
type TryData interface {
	// Decode gives a human readable description of TryData content
	Decode() string
}

//Fun is the common  cost function interface for manipulating TryData
type Fun interface {
	// creates an empty try data store for the decoded part of the try
	CreateData() TryData
	// returns a default parameter which should satisfy constraints
	DefaultParam() *big.Int
	// copies try data from src to dest
	CopyData(dest, src TryData)
	// maximum number of bits used in the try parameter
	MaxLen() int
	//description of the function
	About() string
	//attempts to update hint to give a constraint satisfying try parameter possibly with the help of pre which is assumed to be constraint satisfying. Note it is rare that pre is needed and pre should be un modified by this function
	Constraint(pre TryData, hint *big.Int) (valid bool)
	// Delete hints to the function to remove/replace the ith item
	Delete(i int) bool

	// Idecode decodes the parameter z and stores the result in data.
	IDecode(data TryData, z *big.Int)
}
