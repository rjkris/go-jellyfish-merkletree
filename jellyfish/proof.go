package jellyfish

import "go-jellyfish-merkletree/common"

type SparseMerkleProof struct {
	leaf LeafNode
	siblings []common.HashValue
}

type SparseMerkleRangeProof struct {
	rightSiblings []common.HashValue
}
//
//func (sm *SparseMerkleProof)new(leaf common.SparseMerkleLeafNode, siblings []common.HashValue) SparseMerkleProof {
//	return SparseMerkleProof{
//		leaf:     leaf,
//		siblings: siblings,
//	}
//}
//
//func (smr *SparseMerkleRangeProof)new(rightSiblings []common.HashValue) SparseMerkleRangeProof {
//	return SparseMerkleRangeProof{rightSiblings}
//}
