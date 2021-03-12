package jellyfish

import (
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
	t.Logf("random value: %v", common.HashValue{}.Random())
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
	t.Logf("newRootHash: %+v \n", newRootHash)
	t.Logf("batch: %+v \n", batch)
	for k, v := range batch.NodeBch {
		t.Logf("k: %+v, v: %+v", k, v)
	}
	db.writeTreeUpdateBatch(batch)
	t.Logf("db: %+v \n", db)
	actual, proof := tree.getWithProof(key, 0)
	t.Logf("actual: %+v \n", actual)
	t.Logf("proof: %+v \n", proof)
	assert.Equal(t, value, actual)
}

func TestInsertToPreGenesis(t *testing.T)  {
	db := MockTreeStore{}.new()
	key1 := common.HashValue{0x00}
	value1 := valueT{[]byte{34, 45, 56, 67}}
	preGenesisRootKey := NodeKey{}.newEmptyPath(PreGenesisVersion)
	db.putNode(preGenesisRootKey, LeafNode{}.newLeaf(key1, value1))

	t.Logf("key1: %+v \n", key1)
	tree := JfMerkleTree{db, nil}
	key2 := updateNibble(key1, 0, 15)
	t.Logf("key2: %+v \n", key2)
	value2 := valueT{[]byte{12, 23, 34, 45}}
	_, batch := tree.PutValueSet([]valueSetItem{{key2, value2}}, 0)
	t.Logf("batch: %+v \n", batch)
	assert.Equal(t, 1, batch.StaleNodeIndexBch.Cardinality())
	db.writeTreeUpdateBatch(batch)
	t.Logf("db: %+v \n", db)
	assert.Equal(t, 4, db.numNodes())
	actual1, proof1 := tree.getWithProof(key1, 0)
	acturl2, proof2 := tree.getWithProof(key2, 0)
	t.Logf("proof1: %+v \n", proof1)
	t.Logf("proof2: %+v \n", proof2)
	assert.Equal(t, actual1, value1)
	assert.Equal(t, acturl2, value2)
	t.Logf("stalenodeindex: %+v \n", db.staleNode.String())
	db.purgeStaleNodes(0)
	t.Logf("db: %+v \n", db)
	t.Logf("stalenodeindex: %+v \n", db.staleNode.String())
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
	t.Logf("db1: %+v \n", db)
	actual1, _ := tree.getWithProof(key1, 0)
	assert.Equal(t, value1, actual1)

	// 2. Insert at the previous leaf node. Should generate a branch node at root.
	// Change the 2nd nibble to 1.
	key2 := updateNibble(key1, 1, 1)
	t.Logf("key2: %+v", key2)
	value2 := valueT{[]byte{3, 4}}
	_, batch2 := tree.PutValueSet([]valueSetItem{{key2, value2}}, 1)
	t.Logf("db len: %+v \n", db.numNodes())
	t.Logf("batch2: %+v \n", batch2)
	db.writeTreeUpdateBatch(batch2)
	t.Logf("db2: %+v \n", db)
	actual2, _ := tree.getWithProof(key2, 1)
	assert.Equal(t, value2, actual2)
	assert.Equal(t, 5, db.numNodes())

	t.Log("debug 000000000000000 \n")
	newNp, _ := NibblePath{}.newOdd([]byte{00})
	internalNodeKey := NodeKey{
		Vs: 1,
		np: *newNp,
	}
	t.Log("debug 11111111111111 \n")
	leaf1 := LeafNode{}.newLeaf(key1, value1)
	leaf2 := LeafNode{}.newLeaf(key2, value2)
	children := map[Nibble]Child{}
	children[Nibble(0)] = Child{leaf1.hash(), 1, true}
	children[Nibble(1)] = Child{leaf2.hash(), 1, true}
	internal := InternalNode{}.new(children)

	children = map[Nibble]Child{}
	children[Nibble(0)] = Child{internal.hash(), 1, false}
	rootInternal := InternalNode{}.new(children)
	t.Logf("db: %+v \n", db)
	assert.Equal(t, leaf1, db.getNode(NodeKey{}.newEmptyPath(0)))
	assert.Equal(t, leaf1, db.getNode(internalNodeKey.genChildNodeKey(1, Nibble(0))))
	//t.Logf("internalnodekey: %+v \n", *internalNodeKey.genChildNodeKey(1, Nibble(0)))
	assert.Equal(t, leaf2, db.getNode(internalNodeKey.genChildNodeKey(1, Nibble(1))))
	assert.Equal(t, internal, db.getNode(internalNodeKey))
	assert.Equal(t, rootInternal, db.getNode(NodeKey{}.newEmptyPath(1)))

	// 3. Update leaf2 with new value
	value2Update := valueT{[]byte{5, 6}}
	_, batch3 := tree.PutValueSet([]valueSetItem{{key2, value2Update}}, 2)
	t.Logf("batch3: %+v \n", batch3)
	t.Logf("db len: %+v", db.numNodes())
	db.writeTreeUpdateBatch(batch3)
	t.Logf("db3: %+v \n", db)
	t.Logf("getwithproof debug--------- \n")
	actual3, _ := tree.getWithProof(key2, 0)
	actual4, _ := tree.getWithProof(key2, 1)
	actual5, _ := tree.getWithProof(key2, 2)
	actual6, _ := tree.getWithProof(key1, 2)
	actual7, _ := tree.getWithProof(key2, 2)
	t.Logf("actual3: %+v", actual3)
	t.Logf("actual4: %+v", actual4)
	assert.Nil(t, actual3)
	assert.Equal(t, value2, actual4)
	assert.Equal(t, value2Update, actual5)
	assert.Equal(t, 8, db.numNodes())

	db.purgeStaleNodes(1)
	assert.Equal(t, 7, db.numNodes())
	db.purgeStaleNodes(2)
	assert.Equal(t, 4, db.numNodes())
	assert.Equal(t, value1, actual6)
	assert.Equal(t, value2Update, actual7)
}