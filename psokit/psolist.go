package psokit

import (
	"fmt"
	"log"
	"sort"

	"github.com/mathrgo/setpso"
)

/*
SelectPso primes man to consider using the named SPSO instance creator, name,
for the runs. It checks that there exists an instance creator and returns an
error if it does not exist.There is no need to add a PSO using AddPso if you are
using an inbuilt instance creator name.
*/
func (man *ManPso) SelectPso(name string) error {
	if man.psod[name] == "" {
		return fmt.Errorf("the SPSO instance %s could not be found", name)
	}
	man.psoCase = name
	return nil
}

/*
CreatePso  returns the SPSO instance based on its name and is called by man at
the beginning of a run, so there is no need to call this if using man to execute
the run sequence;  you can use SelectPso() instead.

If SPSO is not found the the event is logged and it returns nil otherwise it
sets up man to use it . This also assumes the cost-function has been created for
man beforehand using CreateFun().
*/
func (man *ManPso) CreatePso(name string) (p PsoInterface) {
	p0 := setpso.NewPso(man.npart, man.f, man.psoSeed0+man.psoSeed1*int64(man.runid))
	switch name {
	case "gpso-0":
		p = setpso.NewGPso(p0)
	case "clpso-0":

		p = setpso.NewCLPso(p0)
	default:
		pc := man.addedPso[name]
		if pc != nil {
			p = pc.Create(p0)

		} else {
			p = nil
			log.Printf("PSO %s not found", name)
		}
	}
	if p != nil {
		man.p = p
		man.psoCase = name
	}
	return
}

// this is done here to give easy comparison with the above list.

/*
loadPsoDescription loads the description of the installed PSO instance.
this is done during the initialisation of man in New().
*/
func (man *ManPso) loadPsoDescription() {

	man.psod = map[string]string{
		"gpso-0":  "single group with global best target; using setpso.NewGPso",
		"clpso-0": "basic comprehensive learning each particle has its own group; using setpso.NewCLPso "}
}

/*
PsoDescription gives a description of SPSOs by name.
*/
func (man *ManPso) PsoDescription() string {
	keys := make([]string, len(man.psod))
	i := 0
	for k := range man.psod {
		keys[i] = k
		i++
	}
	s := fmt.Sprintln("SPSO Description:")
	sort.Strings(keys)
	for i := range keys {
		k := keys[i]
		s += fmt.Sprintf("%s :\n  %s\n", k, man.psod[k])
	}
	return s
}
