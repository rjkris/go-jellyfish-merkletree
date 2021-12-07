package storage
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/rjkris/go-jellyfish-merkletree/common"
//	"github.com/rjkris/go-jellyfish-merkletree/jellyfish"
//	"testing"
//)
//
//func TestLeveldb(t *testing.T) {
//	levelDB, err := New("test Db")
//	if err != nil {
//		t.Logf("new leveldb error: %s", err)
//	}
//	err = levelDB.Put([]byte("leveldb"), []byte("hello leveldb22"))
//	if err != nil {
//		t.Logf("put error: %s", err)
//	}
//	value, _ := levelDB.Get([]byte("leveldb"))
//	t.Log(string(value))
//	if err != nil {
//		t.Logf("get error: %s", err)
//	}
//}
//
//func TestNodeStore(t *testing.T)  {
//	levelDB, _ := New("test Db")
//	leafNode := jellyfish.LeafNode{
//		AccountKey: common.HashValue{}.Random(),
//		ValueHash:  common.HashValue{}.Random(),
//		Value:      "helloword",
//	}
//	children := jellyfish.Children{}
//	children[jellyfish.Nibble(1)] = jellyfish.Child{
//		Hash: common.HashValue{}.Random(),
//		Vs:   0,
//	}
//	internalNode := jellyfish.InternalNode{Children: children}
//	leafJson, _ := json.Marshal(leafNode)
//	internalJson, _ := json.Marshal(internalNode)
//	leafKey := jellyfish.NodeKey{
//		Vs: 0,
//		Np: jellyfish.NibblePath{
//			NumNibbles: 0,
//			Bytes:      [32]uint8{},
//		},
//	}
//	internalKey := jellyfish.NodeKey{
//		Vs: 0,
//		Np: jellyfish.NibblePath{
//			NumNibbles: 2,
//			Bytes:      [32]uint8{},
//		},
//	}
//	leafKeyJson, _ := json.Marshal(leafKey)
//	internalKeyJson, _ := json.Marshal(internalKey)
//	_ = levelDB.Put(leafKeyJson, leafJson)
//	_ = levelDB.Put(internalKeyJson, internalJson)
//	v1, err := levelDB.Get([]byte("aaaaaaaa"))
//	if err != nil {
//		panic(err)
//	}
//	v2, _ := levelDB.Get(internalKeyJson)
//	var newLeaf jellyfish.LeafNode
//	var newInternal jellyfish.InternalNode
//	err = json.Unmarshal(v1, &newInternal)
//	if err != nil {
//		fmt.Println(err)
//	}
//	err = json.Unmarshal(v2, &newLeaf)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Printf("%+v\n", newLeaf)
//	fmt.Printf("%+v\n", newInternal)
//}
