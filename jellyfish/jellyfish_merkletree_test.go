package jellyfish

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go-jellyfish-merkletree/common"
	"testing"
)

type valueT struct {
	value []byte
}

func (v valueT)getValue() []byte {
	return v.value
}

func TestJfMerkleTree_PutValueSet(t *testing.T) {
	fmt.Println(common.HashValue{}.Random())
}

func TestInsertToEmptyTree(t *testing.T)  {
	db := MockTreeStore{}
	tree := JfMerkleTree{
		reader: db,
		value:  nil,
	}
	key := common.HashValue{}.Random()
	value := valueT{[]byte{43, 43, 67, 98}}
	testItem := valueSetItem{key, value}
	newRootHash, batch := tree.PutValueSet([]valueSetItem{testItem}, 0)
	assert.NotEmpty(t, batch)
	fmt.Printf("newRootHash: %+v \n", newRootHash)
	fmt.Printf("batch: %+v \n", batch)
}