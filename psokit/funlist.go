package psokit

import (
	"fmt"
	"log"
	"math/big"
	"sort"

	"github.com/mathrgo/setpso/fun/simplefactor"
	"github.com/mathrgo/setpso/fun/subsetsum"
)

/*
SelectFun informs man to consider using the named cost-function instance, name.
it checks that there exists an instance creator and returns an error if it does
not exist. There is no need to add a cost-function if you are using an
inbuilt instance name.
*/
func (man *ManPso) SelectFun(name string) error {
	if man.fund[name] == "" {
		return fmt.Errorf("the cost-function creator instance %s could not be found", name)
	}
	man.funCase = name
	return nil
}

/*
CreateFun  returns the cost-function instance based on its name and is called by
man at the beginning of each run, so there is no need to call this if using man
to execute the run sequence;  you can use SelectFun() instead to prime man
before a run.

If cost-function not found the the event is logged and it returns nil otherwise
it sets up man to use the cost-function.
*/
func (man *ManPso) CreateFun(name string) (f Fun) {
	// calculate  cost function seed for the run
	fsd := man.funSeed1*int64(man.runid) + man.funSeed0
	switch name {
	case "subsetsum-0":
		// basic subset sum case
		f = subsetsum.New(100, 20, fsd)
	case "simplefactor-30":
		// use this to show that the prime factorisation is still not easy
		var p, q,pMin big.Int
		p.SetString("1059652519", 0)
		q.SetString("929636291", 0)
		pMin.SetString("50000000",0)
		f = simplefactor.New(&p, &q, &pMin)
	case "simplefactor-25":
		var p, q ,pMin big.Int
		p.SetString("30158671", 0)
		q.SetString("26919701", 0)
		pMin.SetString("500000",0)
		f = simplefactor.New(&p, &q,&pMin)
	case "simplefactor-16":
		var p, q, pMin big.Int
		p.SetString("51647", 0)
		q.SetString("97859", 0)
		pMin.SetString("20000",0)
		f = simplefactor.New(&p, &q,&pMin)
	default:
		fc := man.addedFun[name]
		if fc != nil {
			f = fc.Create(fsd)
		} else {
			f = nil
			log.Printf("Cost function creator %s not found", name)
			return
		}
	}
	if f != nil {
		man.f = f
		man.funCase = name
	}
	return
}

/*
this is done during the initialization of man in New().
it is done here to give easy comparison with the above list
*/
func (man *ManPso) loadFunDescription() {

	man.fund = map[string]string{
		"subsetsum-0":     "basic subset sum case 100 elements with up to 20 bit int",
		"simplefactor-30": "30 bit prime factorisation",
		"simplefactor-25": "25 bit prime factorisation",
		"simplefactor-16": "16 bit prime factorisation"}
}

/*
FunDescription gives a description of cost-function by name.
*/
func (man *ManPso) FunDescription() string {
	keys := make([]string, len(man.fund))
	i := 0
	for k := range man.fund {
		keys[i] = k
		i++
	}
	s := fmt.Sprintln("Cost-function Description:")
	sort.Strings(keys)
	for i := range keys {
		k := keys[i]
		s += fmt.Sprintf("%s :\n  %s\n", k, man.fund[k])
	}
	return s
}
