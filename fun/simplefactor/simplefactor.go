// Package simplefactor is  a cost function for finding prime factor pair:
// used to check that it is too difficult for the PSO to
// find the primes.
package simplefactor

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso"
	"github.com/mathrgo/setpso/fun/futil"
)

// Fun type stores data for tests of a factor of the product pq
// Giving a cost of 0 if found.
type Fun struct {
	//NBit is the number of bits used  for the small factor
	Nbit int
	pq   *big.Int
	p    *big.Int
	q    *big.Int
}

//Try is the try interface used by setpso
type Try = setpso.Try

//FunTry gives the try structure to use
type FunTry = futil.IntTry

//TryData is the interface for FunTryData used in package futil
type TryData = futil.TryData

//FunTryData is the decoded data structure for a try
type FunTryData struct {
	p, q, c *big.Int
}

//IDecode decodes z into d
func (f *Fun) IDecode(data TryData, z *big.Int) {
	d := data.(*FunTryData)
	d.p.Set(z)
	d.q.DivMod(f.pq, d.p, d.c)
}

// Decode requests the function to give a meaningful interpretation of
// p as a Parameters subset for the function assuming p satisfies constraints
func (d *FunTryData) Decode() string {
	return fmt.Sprintf("p=%v \nq=%v", d.p, d.q)
}

//IntFunStub gives interface to setpso
type IntFunStub = futil.IntFunStub

// New creates a new function where p and q are the two prime components
// that make up the integer to be factorized
func New(p, q *big.Int) *IntFunStub {
	var f Fun
	f.pq = big.NewInt(0)
	f.p = big.NewInt(0)
	f.p.Set(p)
	f.q = big.NewInt(0)
	f.q.Set(q)
	f.pq.Mul(p, q)
	f.Nbit = p.BitLen()

	return futil.NewIntFunStub(&f)
}

//CreateData creates a empty structure for decoded try
func (f *Fun) CreateData() TryData {
	t := new(FunTryData)
	t.p = new(big.Int)
	t.q = new(big.Int)
	t.c = new(big.Int)
	return t
}

//Cost returns the remainder after dividing p in to the prime product
func (f *Fun) Cost(data TryData, cost *big.Int) {
	cost.Set(data.(*FunTryData).c)
}

//DefaultParam gives a default that satisfies constraints
func (f *Fun) DefaultParam() *big.Int {
	return big.NewInt(2001)
}

//CopyData copies src to dest
func (f *Fun) CopyData(dest, src TryData) {
	s := src.(*FunTryData)
	d := dest.(*FunTryData)
	d.p.Set(s.p)
	d.q.Set(s.q)
	d.c.Set(s.c)
}

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return f.Nbit
}

//Constraint attempts to constrain hint possibly using a copy of pre to do this
func (f *Fun) Constraint(pre TryData, hint *big.Int) (valid bool) {
	if hint.Cmp(big.NewInt(2000)) > 0 {
		valid = true
		if hint.Bit(0) == 0 {
			hint.SetBit(hint, 0, 1)
		}
	} else {
		valid = false
	}
	return
}

// About returns a string description of the contents of Fun
func (f *Fun) About() string {
	var s string
	s = "simple factorise problem\n"
	s += fmt.Sprintf("pq= %v\n", f.pq)
	s += fmt.Sprintf("p= %v\n", f.p)
	s += fmt.Sprintf("q= %v\n", f.q)
	s += fmt.Sprintf("number of bits %d\n", f.Nbit)
	return s
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }

/*Creator is used by psokit to create instances of Fun through its interface
method Create().
*/
type Creator struct {
	p, q *big.Int
}

// NewCreator just returns a Creator of Fun with primes p,q.
func NewCreator(p, q *big.Int) *Creator {
	c := Creator{p, q}
	return &c
}

// Create creates an instance
func (c *Creator) Create(sd int64) setpso.Fun {
	return New(c.p, c.q)
}
