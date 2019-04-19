/*
Package setpso is a collection of Set based Particle Swarm Optimisers(SPSO)
designed for cost functions that map binary patterns to *big.Int cost values.
The binary patterns called Parameters is encoded also as a *big.Int. The SPSO is
a swarm of entities called Particles that together iteratively hunt for better
solutions. The update iteration of the swarm mimics the spirit of the continuous
case and is based on set operations. It also includes experimental enhancements
to improve the discrete case. For brief introduction, context of use and planned
future development read the Readme  file at https://github.com/mathrgo/setpso

Relation to Other Sub Packages

 Package setpso lives in a directory that is at the top of a a hierarchy of
 packages.

 Package setpso contains two working SPSOs: GPso and CLPso that depend on Pso
 for all interfaces except Update() needed in package psokit.

 Packages in setpso/fun is where cost-functions that interface with Pso are
 usually placed and includes any helper packages for such cost-functions.

 Package psokit enables a high level multiple run interface where elements for
 the rum are referred by name to be used in setting up runs of various SPSOs and
 cost-function combinations and searching for good heuristics.

 Particle's Personal-best

 While exploring Parameters for finding reduced cost as returned by the
 independent cost function the Particle keeps a record of the personal best
 Parameter achieved so far called Personal-best with a corresponding best cost.
 The Personal-best status is checked after each update.

Particle's Update Velocity

It represents update Velocity as a vector of weights of the probability of
flipping the corresponding bit at the update iteration. At the beginning of the
update  the velocity is calculated  without flipping bits and then the bits are
flipped with  a probability given by the computed velocity component.  During
the update, once the bit has been flipped the corresponding probability is set
to zero thus avoiding flipping back and keeping the velocity as a vector of
flips that are requested with a given probability to move from a given position
to a desired one that may improve performance.

during the calculation of the velocity of a particle probabilities are combined
using  an operation called pseudo adding where by default probabilities p,q  are
pseudo added to give p+q-pq. Alternatives may be considered in the future such
as max(p,q) if only to show which is best.

Grouping of particles

The particles are split up into groups with each group containing its own
heuristic settings and a list of particles called Targets for it to tend to move
towards. each Particle in the group also uses its Personal-best to move towards.
Various strategies for targeting other Particle's personal-best Parameters or
adapting heuristics can be explored: Pso is not used by its self, since it has
no targets, but forms most common interfaces and has a function PUpdate() that
does the common velocity update. To create a functioning SPSO extra code is
added before PUpdate() to choose Targets and Heuristics which are added by the
derived working SPSOs to generate the total update iteration function, Update().
GPso and CLPso are examples of such derived working SPSOs.

It is important to note that the collection of groups is stored as mapping from
strings  to pointers to groups so groups can be accessed by name  if necessary
although each particle knows which group it belongs to  without using the name
reference. Also groups can have no particles that  belong to the group. At start
up there is only one group called "root" which contains all the particles.
Additional groups  can be formed during initialisation or even during iteration
and particles moved  between groups as and when required.

setpso can be used in low level coding and the higher level run management is provided
by the psokit toolkit package in
    import "github.com/mathrgo/setpso/psokit"
you can quickly get to run an example by going to the setpso/example/runkit1
directory in a terminal then execute
    go run runkit1.go -nrun 1

*/
package setpso
