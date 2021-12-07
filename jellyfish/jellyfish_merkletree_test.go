package jellyfish

import (
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

type testKV struct {
	key common.HashValue
	value ValueT
}

type testKVU struct {
	key common.HashValue
	value ValueT
	updatedValue ValueT
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

func TestJfMerkleTree_PutValueSet(t *testing.T) {
	t.Logf("random Value: %v", common.HashValue{}.Random())
}

func TestInsertToEmptyTree(t *testing.T)  {
	db := MockTreeStore{}.New()
	tree := JfMerkleTree{db, nil}
	key := common.HashValue{}.Random()
	value := ValueT{[]byte{43, 43, 67, 98}}
	testItem := ValueSetItem{key, value}
	newRootHash, batch := (&tree).PutValueSet([]ValueSetItem{testItem}, 0)
	assert.NotEmpty(t, batch)
	assert.NotEmpty(t, batch.StaleNodeIndexBch)
	t.Logf("newRootHash: %+v \n", newRootHash)
	t.Logf("batch: %+v \n", batch)
	for k, v := range batch.NodeBch {
		t.Logf("k: %+v, v: %+v", k, v)
	}
	_ = db.WriteTreeUpdateBatch(batch)
	t.Logf("Db: %+v \n", db)
	actual, proof := tree.getWithProof(key, 0)
	t.Logf("actual: %+v \n", actual)
	t.Logf("proof: %+v \n", proof)
	assert.Equal(t, value, actual)
}

func TestInsertToPreGenesis(t *testing.T)  {
	db := MockTreeStore{}.New()
	key1 := common.HashValue{0x00}
	value1 := ValueT{[]byte{34, 45, 56, 67}}
	preGenesisRootKey := NodeKey{}.newEmptyPath(PreGenesisVersion)
	db.putNode(preGenesisRootKey, LeafNode{}.newLeaf(key1, value1))

	t.Logf("key1: %+v \n", key1)
	tree := JfMerkleTree{db, nil}
	key2 := updateNibble(key1, 0, 15)
	t.Logf("key2: %+v \n", key2)
	value2 := ValueT{[]byte{12, 23, 34, 45}}
	_, batch := tree.PutValueSet([]ValueSetItem{{key2, value2}}, 0)
	t.Logf("batch: %+v \n", batch)
	assert.Equal(t, 1, batch.StaleNodeIndexBch.Cardinality())
	db.WriteTreeUpdateBatch(batch)
	t.Logf("Db: %+v \n", db)
	assert.Equal(t, 4, db.numNodes())
	actual1, proof1 := tree.getWithProof(key1, 0)
	acturl2, proof2 := tree.getWithProof(key2, 0)
	t.Logf("proof1: %+v \n", proof1)
	t.Logf("proof2: %+v \n", proof2)
	assert.Equal(t, actual1, value1)
	assert.Equal(t, acturl2, value2)
	t.Logf("stalenodeindex: %+v \n", db.staleNode.String())
	db.purgeStaleNodes(0)
	t.Logf("Db: %+v \n", db)
	t.Logf("stalenodeindex: %+v \n", db.staleNode.String())
	assert.Equal(t, 3, db.numNodes())
}

func TestInsertAtLeafWithMultipleInternalsCreated(t *testing.T)  {
	db := MockTreeStore{}.New()
	tree := JfMerkleTree{db, nil}

	// 1. Insert the first leaf into empty tree
	key1 := common.HashValue{0x00}
	value1 := ValueT{[]byte{1, 2}}
	_, batch1 := tree.PutValueSet([]ValueSetItem{{key1, value1}}, 0)
	db.WriteTreeUpdateBatch(batch1)
	t.Logf("db1: %+v \n", db)
	actual1, _ := tree.getWithProof(key1, 0)
	assert.Equal(t, value1, actual1)

	// 2. Insert at the previous leaf node. Should generate a branch node at root.
	// Change the 2nd nibble to 1.
	key2 := updateNibble(key1, 1, 1)
	t.Logf("key2: %+v", key2)
	value2 := ValueT{[]byte{3, 4}}
	_, batch2 := tree.PutValueSet([]ValueSetItem{{key2, value2}}, 1)
	t.Logf("Db len: %+v \n", db.numNodes())
	t.Logf("batch2: %+v \n", batch2)
	db.WriteTreeUpdateBatch(batch2)
	t.Logf("db2: %+v \n", db)
	actual2, _ := tree.getWithProof(key2, 1)
	assert.Equal(t, value2, actual2)
	assert.Equal(t, 5, db.numNodes())

	t.Log("debug 000000000000000 \n")
	//newNp, _ := NibblePath{}.newOdd([]byte{00})
	//internalNodeKey := NodeKey{
	//	Vs: 1,
	//	Np: *newNp,
	//}
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
	t.Logf("Db: %+v \n", db)
	//assert.Equal(t, leaf1, Db.getNode(NodeKey{}.newEmptyPath(0)))
	//assert.Equal(t, leaf1, Db.getNode(internalNodeKey.genChildNodeKey(1, Nibble(0))))
	////t.Logf("internalnodekey: %+v \n", *internalNodeKey.genChildNodeKey(1, Nibble(0)))
	//assert.Equal(t, leaf2, Db.getNode(internalNodeKey.genChildNodeKey(1, Nibble(1))))
	//assert.Equal(t, internal, Db.getNode(internalNodeKey))
	actualNode, _ := db.getNode(NodeKey{}.newEmptyPath(1))
	assert.Equal(t, rootInternal, actualNode)

	// 3. Update leaf2 with New Value
	value2Update := ValueT{[]byte{5, 6}}
	_, batch3 := tree.PutValueSet([]ValueSetItem{{key2, value2Update}}, 2)
	t.Logf("batch3: %+v \n", batch3)
	t.Logf("Db len: %+v", db.numNodes())
	db.WriteTreeUpdateBatch(batch3)
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

func TestBatchInsertion(t *testing.T)  {
	key1 := common.HashValue{00}
	value1 := ValueT{[]byte{1}}

	key2 := updateNibble(key1, 0, 2)
	value2 := ValueT{[]byte{2}}
	value2Update := ValueT{[]byte{22}}

	key3 := updateNibble(key1, 1, 3)
	value3 := ValueT{[]byte{3}}

	key4 := updateNibble(key1, 1, 4)
	value4 := ValueT{[]byte{4}}

	key5 := updateNibble(key1, 5, 5)
	value5 := ValueT{[]byte{5}}

	key6 := updateNibble(key1, 3, 6)
	value6 := ValueT{[]byte{6}}

	var batches [][]ValueSetItem
	var oneBatch []ValueSetItem
	batches = append(batches, []ValueSetItem{{key1, value1}})
	batches = append(batches, []ValueSetItem{{key2, value2}})
	batches = append(batches, []ValueSetItem{{key3, value3}})
	batches = append(batches, []ValueSetItem{{key4, value4}})
	batches = append(batches, []ValueSetItem{{key5, value5}})
	batches = append(batches, []ValueSetItem{{key6, value6}})
	batches = append(batches, []ValueSetItem{{key2, value2Update}})

	t.Logf("batches len: %v", len(batches))
	for i, item := range batches {
		if i == 1 {
			continue
		}
		oneBatch = append(oneBatch, item[0])
	}
	t.Logf("onebatch len: %v", len(oneBatch))

	// insert as one batch
    db := MockTreeStore{}.New()
    tree := JfMerkleTree{db, nil}
	_, batch := tree.PutValueSet(oneBatch, 0)
	db.WriteTreeUpdateBatch(batch)

	assert.Equal(t, 12, db.numNodes())
	for _, item := range oneBatch {
		assert.Equal(t, item.Value, tree.treeGetValue(item.HashK, 0))
	}

	// Insert in multiple batches.
	db = MockTreeStore{}.New()
	tree = JfMerkleTree{db, nil}
	_, batch2 := tree.PutValueSets(batches, 0)
	db.WriteTreeUpdateBatch(batch2)

	for _, item := range oneBatch {
		assert.Equal(t, item.Value, tree.treeGetValue(item.HashK, 6))
	}
	assert.Equal(t, 26, db.numNodes())
	db.purgeStaleNodes(1)
	assert.Equal(t, 25, db.numNodes())
	db.purgeStaleNodes(2)
	assert.Equal(t, 23, db.numNodes())
	db.purgeStaleNodes(3)
	assert.Equal(t, 21, db.numNodes())
	db.purgeStaleNodes(4)
	assert.Equal(t, 18, db.numNodes())
	db.purgeStaleNodes(5)
	assert.Equal(t, db.numNodes(), 14)
	db.purgeStaleNodes(6)
	assert.Equal(t, 12, db.numNodes())
	for _, item := range oneBatch {
		assert.Equal(t, item.Value, tree.treeGetValue(item.HashK, 6))
	}
}

func TestNonExistence(t *testing.T)  {
	db := MockTreeStore{}.New()
	tree := JfMerkleTree{db, nil}
	key1 := common.HashValue{0}
	value1 := ValueT{[]byte{1}}

	key2 := updateNibble(key1, 0, 15)
	value2 := ValueT{[]byte{2}}

	key3 := updateNibble(key1, 2, 3)
	value3 := ValueT{[]byte{3}}

	root, batch := tree.PutValueSet([]ValueSetItem{
		{key1, value1},
		{key2, value2},
		{key3, value3}}, 0)
	db.WriteTreeUpdateBatch(batch)

	assert.Equal(t, value1, tree.treeGetValue(key1, 0))
	assert.Equal(t, value2, tree.treeGetValue(key2, 0))
	assert.Equal(t, value3, tree.treeGetValue(key3, 0))
	assert.Equal(t, 6, db.numNodes())

	// test non-existing nodes.
	// 1. Non-existing node at root node
	nonExistingKey := updateNibble(key1, 0, 1)
	t.Logf("nonKey: %v", nonExistingKey)
	value, proof := tree.getWithProof(nonExistingKey, 0)
	t.Logf("proof siblings: %+v \n", proof.siblings)
	assert.Equal(t, nil, value)
	proof.verify(root, nonExistingKey, nil)

	// 2. Non-existing node at non-root internal node
	nonExistingKey = updateNibble(key1, 1, 15)
	value, proof = tree.getWithProof(nonExistingKey, 0)
	assert.Equal(t, nil, value)
	proof.verify(root, nonExistingKey, nil)

	// 3. Non-existing node at leaf node
	nonExistingKey = updateNibble(key1, 2, 4)
	value, proof = tree.getWithProof(nonExistingKey, 0)
	assert.Equal(t, nil, value)
	proof.verify(root, nonExistingKey, nil)
}

func TestManyKeysGetProofAndVerifyTreeRoot(t *testing.T)  {
	numKeys := 10000

	db := MockTreeStore{}.New()
	tree := JfMerkleTree{db, nil}
	var kvs []ValueSetItem
	for i:=0; i<numKeys; i++ {
		key := common.HashValue{}.Random()
		value := ValueT{common.HashValue{}.Random().Bytes()}
		kvs = append(kvs, ValueSetItem{key, value})
		t.Logf("key: %+v, value: %+v", key, value)
	}
	root, batch := tree.PutValueSet(kvs, 0)
	db.WriteTreeUpdateBatch(batch)
	for _, item := range kvs {
		t.Logf("expect key: %v", item.HashK)
		t.Logf("expect value: %+v", item.Value)
		proofValue, proof := tree.getWithProof(item.HashK, 0)
		t.Logf("actual value: %+v", proofValue)
		assert.Equal(t, item.Value, proofValue)
		res := proof.verify(root, item.HashK, item.Value)
		assert.Equal(t, true, res)
	}
}

func TestManyVersionsGetProofAndVerifyTreeRoot(t *testing.T)  {
	numVersions := 10000
	db := MockTreeStore{}.New()
	tree := JfMerkleTree{db, nil}
	var kvus []testKVU
	var roots []common.HashValue
	for i:=0; i<numVersions; i++ {
		key := common.HashValue{}.Random()
		value := ValueT{common.HashValue{}.Random().Bytes()}
		newValue := ValueT{common.HashValue{}.Random().Bytes()}
		kvus = append(kvus, testKVU{
			key:          key,
			value:        value,
			updatedValue: newValue,
		})
	}
	// insert all keys
	for index, item := range kvus {
		root, batch := tree.PutValueSet([]ValueSetItem{{item.key, item.value}}, Version(index))
		roots = append(roots, root)
		db.WriteTreeUpdateBatch(batch)
	}
	// update Value of all keys
	for index, item := range kvus {
		version := Version(numVersions+index)
		root, batch := tree.PutValueSet([]ValueSetItem{{item.key, item.updatedValue}}, version)
		roots = append(roots, root)
		db.WriteTreeUpdateBatch(batch)
	}

	//for index, item := range kvus {
	//	rand.Seed(time.Now().UnixNano())
	//	randomVersion := index+rand.Intn(numVersions-index)
	//	proofValue, proof := tree.getWithProof(item.key, Version(randomVersion))
	//	assert.Equal(t, item.value, proofValue)
	//	assert.Equal(t, true, proof.verify(roots[randomVersion], item.key, item.value))
	//}
	//
	for index, item := range kvus {
		rand.Seed(time.Now().UnixNano())
		randomVersion := index+numVersions+rand.Intn(numVersions-index)
		proofValue, proof := tree.getWithProof(item.key, Version(randomVersion))
		assert.Equal(t, item.updatedValue, proofValue)
		assert.Equal(t, true, proof.verify(roots[randomVersion], item.key, item.updatedValue))
	}
}

func TestInsertToEmptyTreeLevel(t *testing.T)  {
	db := NewTreeStore()
	tree := JfMerkleTree{db, nil}
	key := common.HashValue{}.Random()
	value := ValueT{[]byte{43, 43, 67, 98}}
	testItem := ValueSetItem{key, value}
	newRootHash, batch := (&tree).PutValueSet([]ValueSetItem{testItem}, 0)
	assert.NotEmpty(t, batch)
	assert.NotEmpty(t, batch.StaleNodeIndexBch)
	//t.Logf("newRootHash: %+v \n", newRootHash)
	//t.Logf("batch: %+v \n", batch)
	//for k, v := range batch.NodeBch {
	//	t.Logf("k: %+v, v: %+v", k, v)
	//}
	err := db.WriteTreeUpdateBatch(batch)
	if err != nil {
		panic(err)
	}
	actual, proof := tree.getWithProof(key, 0)
	//t.Logf("actual: %+v \n", actual)
	//t.Logf("proof: %+v \n", proof)
	assert.Equal(t, value, actual)
	assert.Equal(t, true, proof.verify(newRootHash, key, value))
}

func TestManyKeysGetProofAndVerifyTreeRootLevel(t *testing.T)  {
	numKeys := 10000
	db := NewTreeStore()
	tree := JfMerkleTree{db, nil}
	var kvs []ValueSetItem
	for i:=0; i<numKeys; i++ {
		key := common.HashValue{}.Random()
		value := ValueT{common.HashValue{}.Random().Bytes()}
		kvs = append(kvs, ValueSetItem{key, value})
	}
	root, batch := tree.PutValueSet(kvs, 0)
	err := db.WriteTreeUpdateBatch(batch)
	if err != nil {
		panic(err)
	}
	for _, item := range kvs {
		proofValue, proof := tree.getWithProof(item.HashK, 0)
		assert.Equal(t, item.Value, proofValue)
		res := proof.verify(root, item.HashK, item.Value)
		assert.Equal(t, true, res)
	}
}

func TestManyVersionsGetProofAndVerifyTreeRootLevel(t *testing.T)  {
	numVersions := 10000
	db := NewTreeStore()
	tree := JfMerkleTree{db, nil}
	var kvus []testKVU
	var roots []common.HashValue
	for i:=0; i<numVersions; i++ {
		key := common.HashValue{}.Random()
		value := ValueT{common.HashValue{}.Random().Bytes()}
		newValue := ValueT{common.HashValue{}.Random().Bytes()}
		kvus = append(kvus, testKVU{
			key:          key,
			value:        value,
			updatedValue: newValue,
		})
	}
	// insert all keys
	for index, item := range kvus {
		root, batch := tree.PutValueSet([]ValueSetItem{{item.key, item.value}}, Version(index))
		roots = append(roots, root)
		_ = db.WriteTreeUpdateBatch(batch)
	}
	// update Value of all keys
	for index, item := range kvus {
		version := Version(numVersions+index)
		root, batch := tree.PutValueSet([]ValueSetItem{{item.key, item.updatedValue}}, version)
		roots = append(roots, root)
		_ = db.WriteTreeUpdateBatch(batch)
	}

	//for index, item := range kvus {
	//	rand.Seed(time.Now().UnixNano())
	//	randomVersion := index+rand.Intn(numVersions-index)
	//	proofValue, proof := tree.getWithProof(item.key, Version(randomVersion))
	//	assert.Equal(t, item.value, proofValue)
	//	assert.Equal(t, true, proof.verify(roots[randomVersion], item.key, item.value))
	//}
	//
	for index, item := range kvus {
		rand.Seed(time.Now().UnixNano())
		randomVersion := index+numVersions+rand.Intn(numVersions-index)
		proofValue, proof := tree.getWithProof(item.key, Version(randomVersion))
		assert.Equal(t, item.updatedValue, proofValue)
		assert.Equal(t, true, proof.verify(roots[randomVersion], item.key, item.updatedValue))
	}
}
