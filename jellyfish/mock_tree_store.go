package jellyfish

import mapset "github.com/deckarep/golang-set"

type MockTreeStore struct {
	data map[*NodeKey]Node
	staleNode mapset.Set  // StaleNodeIndex
	allowOverwrite bool
}

func (mts MockTreeStore)getNode(nodeK *NodeKey) (interface{}, error) {
	return mts.data[nodeK], nil
}

func (mts MockTreeStore)getRightMostLeaf() (LeafNode, error) {
	return LeafNode{}, nil
}
