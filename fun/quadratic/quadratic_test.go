package quadratic

import (
	"fmt"
	"math/rand"
)

func ExampleNewSampler() {
	p := NewSampler(10)
	in := make([]float64,3)
	out:= make([]float64,1) 
	fmt.Printf("About:\n%s\n", p.About())
	rnd := rand.New(rand.NewSource(3142))
	for i := 0; i < 16; i++ {
		p.Sample(in, out, rnd)
		fmt.Printf("a= %f b= %f c=%f x= %f\n",in[0],in[1],in[2],out[0])
	}

	//Output:
}