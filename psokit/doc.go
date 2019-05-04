/*
Package psokit is a toolkit  for running a collection of, at the moment,
experimental Set-based Particle Swarm Optimisers (SPSOs) in the package setpso.
These SPSOs are used to find good and often optimal cost minimising subsets
where the cost is a function of subsets of a finite collection of items, where
the cost is a big integer so that in principle large and combinatorial difficult
problems can be looked at. Please read  https://github.com/mathrgo/setpso
to give background information  followed by the godoc documentation.

The main toolkit interface is ManPso; its instance is usually referred to as
man which manages the sequence of runs to get some impression of how the
chosen SPSO performs with the chosen cost function with Actions that are
available for monitoring the runs. All these components
(SPSO,Function,Actions) are chosen by name using various Select methods. There
are short descriptions against each name and the toolkit provides methods for
listing the names and their associated descriptions.

As well as inbuilt components there is an interface to generate new component
creators together with their names and brief descriptions which can be moved to
the inbuilt components at a later date. In this way a list of ready made
components can be built up to deepen the test space for  competing algorithms
and monitoring methods.

An example of its use is given in the setpso subdirectory
    setpso/example/runkit1
This includes the command line option reader Action. When run without arguments
command line option help is displayed together with the list of available Actions.
to do a single run try
    go run runkit1.go -nrun 1
in the runkit1 directory.
*/
package psokit
