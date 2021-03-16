package jellyfish

import (
	"fmt"
	_"fmt"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

type SparseMerkleProof struct {
	leaf common.SparseMerkleLeafNode
	/// All siblings in this proof, including the default ones. Siblings are ordered from the bottom
	/// level to the root level.
	siblings []common.HashValue
}

type SparseMerkleRangeProof struct {
	rightSiblings []common.HashValue
}

func (smp *SparseMerkleProof)verify(expectedRootHash common.HashValue, elementKey common.HashValue, elementValue interface{})  {
	fmt.Printf("verify siblings len: %v", len(smp.siblings))
	if len(smp.siblings) > common.LengthInBits {
		panic("siblings len too long")
	}
	if elementValue != nil {
		elementValue := elementValue.(ValueT)
		if elementKey != smp.leaf.ValueHash {
			panic(fmt.Sprintf("keys do not match. key in proof: %v, expected key: %v",
				smp.leaf.ValueHash, elementKey))
		}
		hash := elementValue.Hash()
		if hash != smp.leaf.ValueHash {
			panic(fmt.Sprintf("value hashes do not match. Value hash in proof: %v, expected value hash: %v",
				smp.leaf.ValueHash, hash))
		}
	} else {
		if elementKey == smp.leaf.ValueHash {
			panic(fmt.Sprintf("Expected non-inclusion proof, but key exists in proof."))
		}
	}
	var currentHash common.HashValue
	if smp.leaf == (common.SparseMerkleLeafNode{}) {
		currentHash = common.BytesToHash([]byte("SPARSE_MERKLE_PLACEHOLDER_HASH"))
	} else {
		currentHash = smp.leaf.Hash()
	}
	bitKey := elementKey.Bytes2Bits()
	fmt.Printf("bitKey: %v", bitKey)
	i, j := 0, len(smp.siblings)-1
	for i<len(smp.siblings) {
		if bitKey[j] == 1 {
			currentHash = common.SparseMerkleInternalNode{smp.siblings[i], currentHash}.Hash()
		} else {
			currentHash = common.SparseMerkleInternalNode{currentHash, smp.siblings[i]}.Hash()
		}
		i += 1
		j -= 1
	}
	if currentHash != expectedRootHash {
		panic(fmt.Sprintf("root hashes do not match. Actual root hash: %v, expected root hash: %v",
			currentHash, expectedRootHash))
	}
	return
}

func (smp *SparseMerkleProof)new(leaf common.SparseMerkleLeafNode, siblings []common.HashValue) SparseMerkleProof {
	return SparseMerkleProof{
		leaf:     leaf,
		siblings: siblings,
	}
}

func (smr *SparseMerkleRangeProof)new(rightSiblings []common.HashValue) SparseMerkleRangeProof {
	return SparseMerkleRangeProof{rightSiblings}
}
