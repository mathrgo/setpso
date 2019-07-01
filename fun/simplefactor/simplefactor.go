// Package simplefactor is  a cost function for finding prime factor pair:
// used to check that it is too difficult for the PSO to
// find the primes.
package simplefactor

import (
	"fmt"
	"math/big"

	"github.com/mathrgo/setpso"
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

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return f.Nbit
}

//Cost returns the remainder after dividing p int the prime product
func (f *Fun) Cost(p *big.Int) *big.Int {
	var c, q big.Int
	q.DivMod(f.pq, p, &c)
	return &c
}

// New creates a new function where p and q are the two prime components
// that make up the integer to be factorised
func New(p, q *big.Int) *Fun {
	var f Fun
	f.pq = big.NewInt(0)
	f.p = big.NewInt(0)
	f.p.Set(p)
	f.q = big.NewInt(0)
	f.q.Set(q)
	f.pq.Mul(p, q)
	f.Nbit = p.BitLen()
	return &f
}

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

// ToConstraint uses the previous parameter pre and the updating hint parameter
// to attempt to produce an update to hint which satisfies solution constraints
// and returns valid = True if succeeds. It returns false if the hint is less
// than or equal to 2000 and converts an even hint to an odd hint
func (f *Fun) ToConstraint(pre, hint *big.Int) (valid bool) {
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

// Decode requests the function to give a meaningful interpretation of
// p as a Parameters subset for the function assuming p satisfies constraints
func (f *Fun) Decode(p *big.Int) (s string) {
	var c, q big.Int
	q.DivMod(f.pq, p, &c)
	return fmt.Sprintf("p=%v \nq=%v", p, &q)
}

// Delete hints to the function to remove/replace the ith item
func (f *Fun) Delete(i int) bool { return false }
