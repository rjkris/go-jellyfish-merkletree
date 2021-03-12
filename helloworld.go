package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

type cat struct {
	name string
}
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

	sam := &student{Name:"sam"}
	fmt.Printf("sam addr: %p \n",sam)
	tom := *sam
	fmt.Printf("tom addr: %p \n",&tom)
	tom.Name = "tom"
	fmt.Printf("%v", sam)

	fmt.Println(common.LeadingZeros(128))
	fmt.Println(common.TrailingZeros(2))
	fmt.Println(1<<15)

	fmt.Printf("hashvalue: %+v", common.HashValue{0x00})
	var bigint int64
	//bigint = 1<<64-1
	fmt.Println(bigint)

	mimi2 := test1()
	fmt.Printf("cat2 address: %p", &mimi2)

	mimi := test2()
	fmt.Printf("cat array2 address: %p", mimi)

	cats := map[int]cat{1: {"mimi"}}
	fmt.Printf("cats address: %p", cats)
	cats2 := cats
	fmt.Printf("cats2 address: %p", cats2)
}

func test1() cat {
	var mimi cat
	mimi = cat{"mimi"}
	fmt.Printf("cat address: %p", &mimi)
	return mimi
}

func test2() *map[int]cat {
	res := map[int]cat{1: {"daju"}}
	fmt.Printf("cat array address: %p", &res)
	return &res
}
