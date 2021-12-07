package jellyfish

import (
	"encoding/json"
	"fmt"
	"github.com/rjkris/go-jellyfish-merkletree/leveldb"
)

type treeStore struct {
	Db *leveldb.KVDatabase
}
// -1: None; 1: leaf; 2: internal
type nodeStore struct {
	NodeValue json.RawMessage
	NodeType  int
}

func NewTreeStore() *treeStore {
	db, _ := leveldb.New("statedb")
	return &treeStore{Db: db}
}

func (ts *treeStore)getNode(nodeK NodeKey) (Node, error) {
	//fmt.Println("leveldbbbbbbbbbbbbb")
	nodeKByte, _ := json.Marshal(nodeK)
	nodeVByte, err :=  ts.Db.Get(nodeKByte)
	if err != nil {  // not exit
		return nil, err
	}
	var nodeV nodeStore
	err = json.Unmarshal(nodeVByte, &nodeV)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("nodev debuggggg: %+v", nodeV)
	switch nodeV.NodeType {
	case -1:
		// fmt.Print("leveldbvalue")
		return NoneNode{}, nil
	case 1:
		var leafValue LeafNode
		_ = json.Unmarshal(nodeV.NodeValue, &leafValue)
		// fmt.Printf("leveldbvalue: %+v", leafValue)
		return leafValue, nil
	case 2:
		var internalValue InternalNode
		_ = json.Unmarshal(nodeV.NodeValue, &internalValue)
		// fmt.Printf("leveldbvalue: %+v", internalValue)
		return internalValue, nil
	}
	return nil, fmt.Errorf("type error: %+v", nodeV)
}

func (ts *treeStore)getRightMostLeaf() LeafNode {
	return LeafNode{}
}

func (ts *treeStore) WriteTreeUpdateBatch(batch TreeUpdateBatch) error {
	for k, v := range batch.NodeBch {
		vByte, _ := json.Marshal(v)
		storeKey, storeValue := k, nodeStore{NodeValue: vByte}
		switch v.(type) {
		case NoneNode:
			// fmt.Println("case nonenode")
			storeValue.NodeType = -1
		case LeafNode:
			// fmt.Println("case leafnode")
			storeValue.NodeType = 1
		case InternalNode:
			// fmt.Println("case internalnode")
			storeValue.NodeType = 2
		}
		// fmt.Printf("storeValue debugggggg: %+v\n", storeValue)
		storeKeyByte, _ := json.Marshal(storeKey)
		storeValueByte, err := json.Marshal(storeValue)
		if err != nil {
			// fmt.Println(err)
		}
		// fmt.Printf("storeValueByte: %v\n", storeValueByte)
		var afterValue nodeStore
		_ = json.Unmarshal(storeValueByte, &afterValue)
		// fmt.Printf("aftervalue: %+v", afterValue)
		err = ts.Db.Put(storeKeyByte, storeValueByte)
		if err != nil {
			return err
		}
	}
	return nil
}
