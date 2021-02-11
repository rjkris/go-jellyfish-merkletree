package main

import (
	"fmt"
	"go-jellyfish-merkletree/jellyfish"
)

func main() {
	fmt.Print("hello world")
	list := []int{1}
	fmt.Println(list[len(list)-1])
	fmt.Println(2 / 8)
	fmt.Println(1 << 1)
	test := []int{1,2,3,4,5}
	fmt.Println(test[2:])
	//res := common.BytesToHash([]byte("placeholder"))
	//fmt.Println(res)
	//fmt.Println(jellyfish.Nibble(12))

	v1, v2 := jellyfish.GetChildAndSiblingHalfStart(jellyfish.Nibble(15), 1)
	fmt.Println(v1, v2)
}
