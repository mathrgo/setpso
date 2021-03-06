Introduction
============

The package `setpso` is a collection of experimental set based  Particle Swarm based Optimizer(SPSO) used for finding a near optimal
binary string where each binary string is associated with a cost, supplied by an external cost-function, that is to be minimized. As such it is a combinatorial optimizer and is based on the continuous swarm optimizer. The cost value can have several representations ranging from big integers to a floating point noisy value depending on the cost to be approximately minimized.

The swarm of _Particles_ iterates towards finding better solutions. Each Particle in the swarm has a candidate exploratory solution as a binary string together with an array of _Velocity_ components and a record of its _Personal-best_ solution so far. Each Particle is guided, through Velocity updates, by its Personal-best and a collection of chosen other Particles, called _Targets_, by using their Personal-best solutions as well.  

Subset Encoding
-----------

The binary string is stored as a big positive integer and is also treated as
a subset where a bit value of 1 in the ith position corresponds to the ith element being a member of the set
and 0 corresponds to it being a non member element of the set associated with the binary string. Following the continuous case by analogy the binary string can be regarded as a vector of parameters that happen to take on the values 0 or 1 and as such it is called the _Parameters_ of the Particle.

Experimental Components
-----------------------

1. Each binary bit has associated velocity which loosely corresponds to the
probability of the bit flipping in the next update, which is set to zero after the update if the flip occurs during the update.

2. prior to the update the velocity is changed to increase the chance of mutation through flipping that is proportional to the Hamming distance to various targets plus a small bias; this mimics the continuous case where such a shotgun approach guarantees finding genuine local minima.

3. velocity probabilities are combined using a second order polynomial:

    _P(p,q)=p+q-pq_

   for combining probabilities p and q. It can be easily checked that as an
   infix operation it is both commutative and associative and combined
   probability is usually greater than the individual probabilities and always in the range 0.0 to 1.0.
   (provided they are both in the range from 0.0 to 1.0).  

4. The swarm is partitioned into _Groups_ of particles with the same Targets and
velocity update parameters called _Heuristics_ .The Target Particles and
Heuristics for the group are not fixed and can be programmed to change prior to
an update. How this is done depends on the variant of SPSO in use.

5. The core of the SPSO is the same for all variants which only differ by how the
Groups are organized.

6. The process of setting up batch runs and testing involves similar processes over
and over again so a tool kit has been produced to simplify this. The tool kit has
a list of cost-functions and SPSO variants and one can build in cases of interest.
Code for the toolkit is in the subdirectory `psokit` .

7. The tool kit has various functions to enable new
components to be introduced as well as new SPSOs and cost-functions. It also includes
a interface to include actions that trigger at specific points of the runs.

8. To support the tool kit all SPSOs will have to meet a golang interface specified
in the `setpso` package. Likewise for all cost-functions and monitoring
components.

Hopefully this will produce a rich class of SPSOs that perform well. The main
reason for producing these novel changes is to find a better class of SPSO.
However, it may not be so. Thus there is a need to produce competing SPSOS to
compare it with. for instance:

* Splitting the Target influence down further to be a function of individual bit
position may work better, although the mutation described in 2. should remove
the need to do this.

* Combining probabilities using the maximum of the two probabilities may work
better. Here I think the combining of probabilities as given in 3. above is
smooth and enables Both probabilities to contribute in more cases. For instance
if p=q=0.5 then the combined probability is 0.75 rather than 0.5.

* Removing the targeting of a Particle partly to its Personal-best may give better
results.

Any help in producing such variants for comparison and possible adoption are always welcome, but are for the moment not done here.

Possible Future Development
---------------------------

1. The current SPSOs in this package do quickly converge to near best solutions and
take significantly more time to hit the optimal solution because of the nature
of directed random search so they will be used more for finding approximate
solutions to be used in machine learning where the process is stopped to avoid
over fitting and there is no need to go further.

2. The SPSO should be a generalist so lots of test cost functions need to be coded
up and SPSO evolved to work reasonably well on these test cost-functions.
However, the heuristics and targeting should be adaptable to the cost-function
at hand by simple tracking of cost and size of sets involved.

3. The test cost-functions should have a large number of functions that try to find
_algorithms_ that fit criteria so the SPSO can be used to find its own
expressions for heuristic values and target choice by using a bootstrap process.

4. Support for a Cost-function  that changes, but infrequently, so the personal
best cost has to be  re-evaluated from time to time. This could for instance
include the Cost-function replacing a set item  that is never used to give a low
cost by another item that was not in the original set of items which may
result in improved cost resulting in changing the items properties. As part of this
The number of set elements should be allowed to change from time to time. 
At the moment this is partially supported through repeated evaluation of costs with
one cost type that supports averaging costs for a given parameter with a forgetting
time constant.

5. If a group ends up with redundancy that is a lot of Particles converge to a
single Subset then some of the Particles could be reassigned to a new
exploratory group to search for better alternative solutions. This could also
mean storing a good subset before removing it and then using it later on.
Perhaps a collection of such good solutions could form a new optimization where
these good solutions are elements of a larger set that encodes how to combine
these good solutions to produce better ones - a crude form of subroutines.  

Getting Started
---------------

golang is a Google programming language that appears to be suited for developing
SPSOs. ensure you have a copy by going to [golang](https://golang.org) and down
loading it.

Sofware is developed using Visual Studio Code at <https://code.visualstudio.com>  which is the preferred IDE.

The following instructions are for quickly getting a working example using a command line at a terminal: 

Open a terminal and clone get this package to your computer by going to an empty directory and typing 

    git clone https://github.com/mathrgo/setpso .

now move to the subdirectory

    example/runkit1

and type

    go run runkit1.go

this will give you a set of command options to use

to do a single default run type

    go run runkit1.go -nrun 1

This may not mean that much at this point but you have  found a solution to the subset sum problem of 100 items with element items randomly chosen from 20 bit positive integers.

Documentation
-------------

For background information and some insight into the algorithms read <https://github.com/mathrgo/setpso/blob/master/doc/psoForML.pdf> .

Code derived documentation is given in <https://godoc.org/github.com/mathrgo/setpso> .

This documentation is hierarchical in nature giving information on each package and is worth reading to understand the packages starting with setpso.
 
follow the instructions and example for running GPso and CLPso .

