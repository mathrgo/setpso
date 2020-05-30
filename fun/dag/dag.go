/*
Package dag  provides the base for Directed Asynchronous Graphs functions with
two inputs per node. It also includes examples using this base.
*/
package dag

import (
	"fmt"
	"math/big"
)

/*
Dag is the method for encoding/decoding special type of Directed Asynchronous
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
type Dag struct {
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
	// number of used nodes
	usedNodes int
	// interface to node operation
	Opt Opt
	// sum of used node costs
	structureCost int
}

/*
Opt is bearbones interface used by Dag to decode operations into human readable form for the nodes and
reserve spacing for the nodes operation encoding.
*/
type Opt interface {
	// gives a description of the operation type
	About() string
	//this gives a human readable  version of the encoded operation opt.
	Decode(opt *big.Int) string
	// Number of bits needed to store opt encoding.
	BitSize() int
	// cost of using opt
	Cost(opt *big.Int) int
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

//The code  for output type from node.
const (
	// output not yet allocated
	OTypeNone = iota
	// output allocated to be used by a node
	OTypeNode
	// output adopted by an output
	OTypeOutput
)

//Node is  interior node Data
type Node struct {
	// the type of the input code from the IType list of costants
	ltype, rtype int
	//type of node output
	otype int
	// the index into each type array
	lindex, rindex int
	// encoded node operation
	opt *big.Int
	// store of opt cost
	optCost int
}

/*
NewDag creates a dag-function base for carrying out common  operations. opt gives
the detailed  process of converting the input pair of an interior node  into an
output. nvar gives the number of inputs to the function; nnode  is the  number
of interior nodes; nout is the number of expected outputs which should be
significantly less than nnode; nbitslookback is the number of bits  used to give
a look-back offset integer, where a look-back offset of i points to the i+1 th
object before the inner-node. Note the inputs are before the inner node and
anything before an input is regarded as a in put with a constant value of 1.

using sampleSize random samples.
*/
func NewDag(nvar, nnode, nout, nbitslookback int, opt Opt) *Dag {
	d := new(Dag)
	d.nVar = nvar
	d.nNode = nnode
	d.nOut = nout
	d.nBitsLookback = nbitslookback
	d.outNodes = make([]int, nout)
	d.INodes = make([]Node, nnode)
	d.Opt = opt
	d.optSize = d.Opt.BitSize()

	// play safe
	for i := range d.INodes {
		d.INodes[i].otype = OTypeNone
		d.INodes[i].opt = big.NewInt(0)
	}
	return d
}

// MaxLen returns the number of elements (bits) for the encoding.
// It is maximum number of bits in the parameter big integer which is also the
// maximum number of elements in the subset
func (d *Dag) MaxLen() int {
	return d.nNode * (2*d.nBitsLookback + d.optSize)
}

// DecodeDag requests the function to give a meaningful interpretation of
// z as a Parameters subset for the function assuming z satisfies constraints
func (d *Dag) DecodeDag(z *big.Int) (s string) {
	if !d.Idecode(z) {
		s = fmt.Sprintf("DAG does not have enough outputs\n")
		return
	}
	s = fmt.Sprintf("DAG structure:\n")
	s += fmt.Sprintf("Used %d nodes\n", d.usedNodes)
	s += fmt.Sprintf("structure cost %d\n", d.structureCost)
	for i := 0; i < d.usedNodes; i++ {
		s += fmt.Sprintf("NODE %d: [ ", i)
		nd := d.INodes[i]
		// add left node input description
		switch nd.ltype {
		case ITypeVar:
			s += fmt.Sprintf("in%d ", nd.lindex)
		case ITypeNode:
			s += fmt.Sprintf("nd%d ", nd.lindex)
		case ITypeConst:
			s += fmt.Sprintf("1 ")
		}
		// add node operation description
		s += fmt.Sprintf("%s ", d.Opt.Decode(nd.opt))
		// add right node input description
		switch nd.rtype {
		case ITypeVar:
			s += fmt.Sprintf("in%d ]=>", nd.rindex)
		case ITypeNode:
			s += fmt.Sprintf("nd%d ]=>", nd.rindex)
		case ITypeConst:
			s += fmt.Sprintf("1 ]=>")
		}
		// add output type
		switch nd.otype {
		case OTypeNone:
			s += "None"
		case OTypeNode:
			s += "Node"
		case OTypeOutput:
			s += "Out"
		}
		s += fmt.Sprintln()
	}
	return

}

/*
Idecode converts  a into a corresponding DAG of interior nodes. Is used  as a
precursor to using it as a algorithm function multiple times to build up  a cost
evaluation. It returns false if it fails to do so.
*/
func (d *Dag) Idecode(a *big.Int) (ok bool) {
	ok = true
	var nodeBase int // is the binary string bit location for node data
	//populate all nodes
	for i := range d.INodes {
		nd := &d.INodes[i]
		// clear node output use
		nd.otype = OTypeNone
		// now pick up the bits to determine node operations
		opt := big.NewInt(0)
		for j := 0; j < d.optSize; j++ {
			opt.SetBit(opt, j, a.Bit(j+nodeBase))
		}
		nd.opt.Set(opt)
		// skip over opt encoding
		nodeBase += d.optSize
		if opt.Sign() == 0 {
			// treat this as an empty node that is not used
			nd.ltype = ITypeConst
			nd.lindex = 0
			nd.rtype = ITypeConst
			nd.rindex = 0
		} else {
			//l,r are look-back offsets for node input
			// make sure nodes have previous elements as input
			l := 1
			r := 1
			pwr2 := 1 //power of 2 factor to convert bits to look-back integers
			for j := 0; j < d.nBitsLookback; j++ {
				jl := j + nodeBase
				jr := jl + d.nBitsLookback
				if a.Bit(jl) == 1 {
					l += pwr2
				}
				if a.Bit(jr) == 1 {
					r += pwr2
				}
				pwr2 <<= 1
			}
			//fmt.Printf("l=%d r=%d\n", l, r)

			// process left input to node
			l = i - l   // convert to absolute node position
			if l >= 0 { // we have a previous node  as input
				nd.ltype = ITypeNode
				nd.lindex = l
				d.INodes[l].otype = OTypeNode
			} else {
				l := -l
				if l <= d.nVar { // we have a global variable input
					// notice indexing is backwards
					nd.ltype = ITypeVar
					nd.lindex = l - 1
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
				d.INodes[r].otype = OTypeNode
			} else {
				r := -r
				if r <= d.nVar { // we have a global variable input
					// notice indexing is backwards
					nd.rtype = ITypeVar
					nd.rindex = r - 1
				} else { // out of range so set to constant 1 input
					nd.rtype = ITypeConst
					nd.rindex = 0
				}
			}

		}
		// skip over lookback encoding
		nodeBase += 2 * d.nBitsLookback
	}
	// now allocate output to the first set of unused non empty INodes it is not
	// clear what to do if not all outputs are allocated for the moment return ok
	// as false indicating a constraint failure
	outLength := d.nOut
	nodesLength := len(d.INodes)
	o := 0
	n := 0
	d.structureCost = 0
	for ; o < outLength && n < nodesLength; n++ {
		nd := &d.INodes[n]
		if nd.otype == OTypeNone && nd.opt.Sign() != 0 {
			nd.otype = OTypeOutput
			d.outNodes[o] = n
			o++
		}
		d.structureCost += d.Opt.Cost(nd.opt)
	}
	d.usedNodes = n
	if o < outLength {
		ok = false
	}
	return
}
