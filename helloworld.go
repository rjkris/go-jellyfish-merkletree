package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"go-jellyfish-merkletree/common"
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

	//v1, v2 := jellyfish.GetChildAndSiblingHalfStart(jellyfish.Nibble(15), 1)
	//fmt.Println(v1, v2)
	//
	//
	//const N = 3
	//
	//m := make(map[int]*int)
	//
	//for i := 0; i < N; i++ {
	//	m[i] = &i
	//	fmt.Println(m[i])
	//}
	//
	//for _, v := range m {
	//	fmt.Println(v)
	//}

	dict := map[int]int{}
	dict[1] = 1
	fmt.Println(dict)
	delete(dict, 1)
	delete(dict, 2)
	fmt.Println(dict)
	fmt.Println(dict)
	fmt.Println(true)
	a := mapset.NewSet()
	fmt.Println(a.Add(1))
	fmt.Println(a.Add(5))
	fmt.Println(a.Add(1))
	type people interface {
		getName()
	}
	type student struct {
		Name string
	}
	dict2 := map[student]student{}
	println(dict[9])
	fmt.Printf("%T", dict2[student{}])

	sam := student{Name:"sam"}
	tom := sam
	tom.Name = "tom"
	fmt.Printf("%v", sam)

	fmt.Println(common.LeadingZeros(128))
	fmt.Println(common.TrailingZeros(2))
	fmt.Println(1<<15)
}