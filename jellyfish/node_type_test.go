package jellyfish

import (
	"github.com/rjkris/go-jellyfish-merkletree/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInternalHashAndProof(t *testing.T)  {
	//bytes := common.HashValue{}.Random().Bytes()
	//bytes[len(bytes)-1] &= 0xf0
	//internalNodeKey := NodeKey{0, *NibblePath{}.new(bytes)}
	children := Children{}
	index1 := Nibble(4)
	index2 := Nibble(15)
	hash1 := common.HashValue{}.Random()
	hash2 := common.HashValue{}.Random()
	//child1NodeKey, _ := genLeafKeys(0, internalNodeKey.np, index1)
	//child2NodeKey, _ := genLeafKeys(1, internalNodeKey.np, index2)
	children[index1] = Child{hash1, 0, false}
	children[index2] = Child{hash2, 1, false}
	internalNode := InternalNode{}.new(children)
	sparseMerklePlaceholderHash := common.BytesToHash([]byte("SPARSE_MERKLE_PLACEHOLDER_HASH"))
	hashX1 := common.SparseMerkleInternalNode{hash1, sparseMerklePlaceholderHash}.Hash()
	hashX2 := common.SparseMerkleInternalNode{hashX1, sparseMerklePlaceholderHash}.Hash()
	hashX3 := common.SparseMerkleInternalNode{sparseMerklePlaceholderHash, hashX2}.Hash()
	hashX4 := common.SparseMerkleInternalNode{sparseMerklePlaceholderHash, hash2}.Hash()
	hashX5 := common.SparseMerkleInternalNode{sparseMerklePlaceholderHash, hashX4}.Hash()
	hashX6 := common.SparseMerkleInternalNode{sparseMerklePlaceholderHash, hashX5}.Hash()
	rootHash := common.SparseMerkleInternalNode{hashX3, hashX6}.Hash()
	assert.Equal(t, internalNode.hash(), rootHash)
}

func genLeafKeys(version Version, nibblePath NibblePath, nibble Nibble) (NodeKey, common.HashValue) {
	(&nibblePath).push(nibble)
	var accountKey common.HashValue
	accountKey.SetBytes(nibblePath.Bytes)
	return NodeKey{version, nibblePath}, accountKey
}
