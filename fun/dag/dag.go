/*
Package dag  provides the base for Directed Asynchronous Graphs functions with
two inputs per node. It also includes examples using this base.
*/
package dag

import "math/big"

/*
Fun is the method for encoding/decoding special type of Directed Asynchronous
Graph into a binary string with an array of input nodes  and an array of ordered
interior nodes.

Node Linking

For encoding/decoding conceptually The inputs are placed before the interior
nodes to make one contiguous  array and each interior node has a pair of offset
values of positive integers each obtained by adding  1 to a binary string of
length nBinOffset. The offset values are used to link each interior  node to two
inputs. if an offset reaches further than available slots the offset is taken to
be pointing to an element with value 1.

Interior Node Operation

Each interior node is allocated OptSize() bits to encode how the two inputs to the node
are operated on to give an output value. At this level how this is done is
opaque: an interface called Opt does this instead using the method Opt().
*/
type Fun struct {
	// number of inputs
	nVar int
	//current number of interior nodes
	nNode int
	// number of outputs
	nOut int
	// maximum number of bits in the look back offset
	nBitsLookback int
	// out put node index array
	outNodes []int
	// number of bits needed to store a node's operations code.
	optSize int
	// interior node array
	INodes []Node
	// interface to node operation
	Opt Opt
}

/*
Opt is  interface for interior node operation. l and r are the left and right
inputs to the node, opt is a, positive integer,  operations code where typically
the the least  significant bits of opt give encoding of the operation such as +
or * and  the remaining bits encode input factors with sign being encoded in
this bit as well.
*/
type Opt interface {
	// function to generate a node output
	Opt(l, r, opt big.Int) big.Int
	// Number of bits needed to store opt
	BitSize() int
}

//The code for the input sources for the node.
const (
	// source from global input array
	ITypeVar = iota
	//  source from another node
	ITypeNode
	// source from a constant node of value 1
	ITypeConst
)

//Node is  interior node Data
type Node struct {
	// the type of the input code from the IType list of costants
	ltype, rtype int
	// the index into each type array
	lindex, rindex int
	// encoded node operation
	opt big.Int
}

/*
New creates a dag-function base for carrying out common  operations. opt gives
the detailed  process of converting the input pair of an interior node  into an
output. nvar gives the number of inputs to the function; nnode  is the  number
of interior nodes; nout is the number of expected outputs which should be
significantly less than nnode; nbitslookback is the number of bits  used to give
a look-back offset integer, where a look-back offset of i points to the i+1 th
object before the inner-node. Note the inputs are before the inner node and
anything before an input is regarded as a in put with a constant value of 1.
*/
func New(nvar, nnode, nout, nbitslookback int, opt Opt) *Fun {
	f := new(Fun)
	f.nVar = nvar
	f.nNode = nnode
	f.nOut = nout
	f.nBitsLookback = nbitslookback
	f.outNodes = make([]int, nout)
	f.INodes = make([]Node, nnode)
	f.Opt = opt
	f.optSize = f.Opt.BitSize()
	return f
}

/*
Idecode converts  a into a corresponding DAG of interior nodes and is used  as a
precursor to using it as a algorithm function multiple times to build up  a cost
evaluation.
*/
func (f *Fun) Idecode(a big.Int) {
	var nodeBase int // is the binary string bit location for node data
	//populate all nodes
	for i := range f.INodes {
		//l,r are look-back offsets for node input
		// make sure nodes have previous elements as input
		l := -1
		r := -1
		pwr2 := 1 //power of 2 factor to convert bits to look-back integers
		for j := 0; j < f.nBitsLookback; j++ {
			jl := j + nodeBase
			jr := jl + f.nBitsLookback
			if a.Bit(jl) == 1 {
				l += pwr2
			}
			if a.Bit(jr) == 1 {
				r += pwr2
			}
			pwr2 <<= 1
		}
		nd := f.INodes[i]
		// process left input to node
		l = i - l   // convert to absolute node position
		if l >= 0 { // we have a previous node  as input
			nd.ltype = ITypeNode
			nd.lindex = l
		} else {
			l := -l
			if l < f.nVar { // we have a global variable input
				// notice indexing is backwards
				nd.ltype = ITypeVar
				nd.lindex = l
			} else { // out of range so set to constant 1 input
				nd.ltype = ITypeConst
				nd.lindex = 0
			}
		}
		// process right input to node
		r = i - r   // convert to absolute node position
		if r >= 0 { // we have a previous node  as input
			nd.rtype = ITypeNode
			nd.rindex = r
		} else {
			r := -r
			if r < f.nVar { // we have a global variable input
				// notice indexing is backwards
				nd.rtype = ITypeVar
				nd.rindex = r
			} else { // out of range so set to constant 1 input
				nd.rtype = ITypeConst
				nd.rindex = 0
			}
		}
		// now pick up the bits to determine node operations
		opt := big.NewInt(0)
		nodeBase += 2 * f.nBitsLookback
		for j := 0; j < f.optSize; j++ {
			opt.SetBit(opt, j, a.Bit(j+nodeBase))
		}
		nd.opt.Set(opt)
		nodeBase += f.optSize
	}
}

// MaxLen returns the number of elements in the subset sum problem
func (f *Fun) MaxLen() int {
	return f.nNode * (2*f.nBitsLookback + f.optSize)
}
