// this is a utility for generating prime pairs
package main

import (
	"crypto/rand"
	"fmt"
)

func genAndPrint(bits int) {
	p, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	fmt.Println(p)
}
func main() {
	bits := 30
	println("bits=", bits)
	genAndPrint(bits)
	genAndPrint(bits)

}

//some results:
//----R1-----
// bits= 25
// 30158671
// 26919701
//----R2-----
// bits= 30
// 1059652519
// 929636291
