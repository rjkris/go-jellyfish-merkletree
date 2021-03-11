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

func updateNibble(originalKey common.HashValue, n uint, nibble uint8) common.HashValue {
	var res common.HashValue
	if nibble > 16 {
		panic("nibble too large")
	}
	key := originalKey.Bytes()
	if n % 2 == 0 {
		key[n/2] = key[n/2] & 0x0f | nibble << 4
	} else {
		key[n/2] = key[n/2] & 0xf0 | nibble
	}
	var byteArray [common.HashLength]byte
	for i, v := range key {
		byteArray[i] = v
	}
	res.SetBytes(byteArray)
	return res
}

func (v valueT)getValue() []byte {
	return v.value
}

func TestJfMerkleTree_PutValueSet(t *testing.T) {
	fmt.Println(common.HashValue{}.Random())
}

func TestInsertToEmptyTree(t *testing.T)  {
	db := MockTreeStore{}.new()
	tree := JfMerkleTree{
		reader: db,
		value:  nil,
	}
	key := common.HashValue{}.Random()
	value := valueT{[]byte{43, 43, 67, 98}}
	testItem := valueSetItem{key, value}
	newRootHash, batch := (&tree).PutValueSet([]valueSetItem{testItem}, 0)
	assert.NotEmpty(t, batch)
	assert.NotEmpty(t, batch.StaleNodeIndexBch)
	fmt.Printf("newRootHash: %+v \n", newRootHash)
	fmt.Printf("batch: %+v \n", batch)
	for k, v := range batch.NodeBch {
		fmt.Printf("k: %+v, v: %+v", k, v)
	}
	db.writeTreeUpdateBatch(batch)
	fmt.Printf("db: %+v \n", db)
	actual, proof := tree.getWithProof(key, 0)
	fmt.Printf("actual: %+v \n", actual)
	fmt.Printf("proof: %+v \n", proof)
	assert.Equal(t, value, actual)
}

func TestInsertToPreGenesis(t *testing.T)  {
	db := MockTreeStore{}.new()
	key1 := common.HashValue{0x00}
	value1 := valueT{[]byte{34, 45, 56, 67}}
	preGenesisRootKey := NodeKey{}.newEmptyPath(PreGenesisVersion)
	db.putNode(*preGenesisRootKey, LeafNode{}.newLeaf(key1, value1))

	fmt.Printf("key1: %+v \n", key1)
	tree := JfMerkleTree{db, nil}
	key2 := updateNibble(key1, 0, 15)
	fmt.Printf("key2: %+v \n", key2)
	value2 := valueT{[]byte{12, 23, 34, 45}}
	_, batch := tree.PutValueSet([]valueSetItem{{key2, value2}}, 0)
	fmt.Printf("batch: %+v \n", batch)
	assert.Equal(t, 1, batch.StaleNodeIndexBch.Cardinality())
	db.writeTreeUpdateBatch(batch)
	fmt.Printf("db: %+v \n", db)
	assert.Equal(t, 4, db.numNodes())
	//actual1, proof1 := tree.getWithProof(key1, 0)
	//acturl2, proof2 := tree.getWithProof(key2, 0)
	//fmt.Printf("proof1: %+v \n", proof1)
	//fmt.Printf("proof2: %+v \n", proof2)
	//assert.Equal(t, actual1, value1)
	//assert.Equal(t, acturl2, value2)
	fmt.Printf("stalenodeindex: %+v \n", db.staleNode.String())
	db.purgeStaleNodes(0)
	fmt.Printf("db: %+v \n", db)
	fmt.Printf("stalenodeindex: %+v \n", db.staleNode.String())
	assert.Equal(t, 3, db.numNodes())
}

func TestInsertAtLeafWithMultipleInternalsCreated(t *testing.T)  {
	db := MockTreeStore{}.new()
	tree := JfMerkleTree{db, nil}

	// 1. Insert the first leaf into empty tree
	key1 := common.HashValue{0x00}
	value1 := valueT{[]byte{1, 2}}
	_, batch1 := tree.PutValueSet([]valueSetItem{{key1, value1}}, 0)
	db.writeTreeUpdateBatch(batch1)
	fmt.Printf("db1: %+v \n", db)
	actual1, _ := tree.getWithProof(key1, 0)
	assert.Equal(t, value1, actual1)

	// 2. Insert at the previous leaf node. Should generate a branch node at root.
	// Change the 2nd nibble to 1.
	key2 := updateNibble(key1, 1, 1)
	fmt.Printf("key2: %+v", key2)
	value2 := valueT{[]byte{3, 4}}
	_, batch2 := tree.PutValueSet([]valueSetItem{{key2, value2}}, 1)
	fmt.Printf("db len: %+v \n", db.numNodes())
	fmt.Printf("batch2: %+v \n", batch2)
	db.writeTreeUpdateBatch(batch2)
	fmt.Printf("db2: %+v \n", db)
	actual2, _ := tree.getWithProof(key2, 1)
	assert.Equal(t, value2, actual2)
	assert.Equal(t, 5, db.numNodes())

	fmt.Println("debug 000000000000000")
	newNp, _ := NibblePath{}.newOdd([]byte{00})
	internalNodeKey := NodeKey{
		Vs: 1,
		np: *newNp,
	}
	fmt.Println("debug 11111111111111")
	leaf1 := LeafNode{}.newLeaf(key1, value1)
	leaf2 := LeafNode{}.newLeaf(key2, value2)
	children := map[Nibble]Child{}
	children[Nibble(0)] = Child{leaf1.hash(), 1, true}
	children[Nibble(1)] = Child{leaf2.hash(), 1, true}
	internal := InternalNode{}.new(children)

	children = map[Nibble]Child{}
	children[Nibble(0)] = Child{internal.hash(), 1, false}
	rootInternal := InternalNode{}.new(children)
	fmt.Printf("db: %+v \n", db)
	assert.Equal(t, leaf1, db.getNode(*NodeKey{}.newEmptyPath(0)))
	assert.Equal(t, leaf1, db.getNode(*internalNodeKey.genChildNodeKey(1, Nibble(0))))
	//fmt.Printf("internalnodekey: %+v \n", *internalNodeKey.genChildNodeKey(1, Nibble(0)))
	assert.Equal(t, leaf2, db.getNode(*internalNodeKey.genChildNodeKey(1, Nibble(1))))
	assert.Equal(t, internal, db.getNode(internalNodeKey))
	assert.Equal(t, rootInternal, db.getNode(*NodeKey{}.newEmptyPath(1)))

	// 3. Update leaf2 with new value
	value2Update := valueT{[]byte{5, 6}}
	_, batch3 := tree.PutValueSet([]valueSetItem{{key2, value2Update}}, 2)
	fmt.Printf("batch3: %+v \n", batch3)
	fmt.Printf("db len: %+v", db.numNodes())
	db.writeTreeUpdateBatch(batch3)
	fmt.Printf("db3: %+v \n", db)
	fmt.Printf("getwithproof debug--------- \n")
	//actual3, _ := tree.getWithProof(key2, 0)
	actual4, _ := tree.getWithProof(key2, 1)
	//actual5, _ := tree.getWithProof(key2, 2)
	//actual6, _ := tree.getWithProof(key1, 2)
	//actual7, _ := tree.getWithProof(key2, 2)
	//fmt.Printf("actual3: %+v", actual3)
	//assert.Nil(t, actual3)
	assert.Equal(t, value2, actual4)
	//assert.Equal(t, value2Update, actual5)
	//assert.Equal(t, 8, db.numNodes())
	//
	//db.purgeStaleNodes(1)
	//assert.Equal(t, 7, db.numNodes())
	//db.purgeStaleNodes(2)
	//assert.Equal(t, 4, db.numNodes())
	//assert.Equal(t, value1, actual6)
	//assert.Equal(t, value2Update, actual7)
}