package jellyfish

import mapset "github.com/deckarep/golang-set"

type MockTreeStore struct {
	data map[NodeKey]Node
	staleNode mapset.Set  // StaleNodeIndex
	allowOverwrite bool
}

func (mts MockTreeStore)new() *MockTreeStore {
	return &MockTreeStore{
		data:           map[NodeKey]Node{},
		staleNode:      mapset.NewSet(),
		allowOverwrite: false,
	}
}
func (mts *MockTreeStore)getNode(nodeK NodeKey) interface{} {
	return mts.data[nodeK]
}

func (mts *MockTreeStore)getRightMostLeaf() LeafNode {
	return LeafNode{}
}

func (mts *MockTreeStore)writeTreeUpdateBatch(batch TreeUpdateBatch)  {
	for k, v := range batch.NodeBch {
		mts.putNode(k, v)
	}
	for k := range batch.StaleNodeIndexBch.Iter() {
		i := k.(StaleNodeIndex)
		mts.putStaleNodeIndex(i)
	}
}

func (mts *MockTreeStore)putNode(nodeK NodeKey, node Node) {
	mts.data[nodeK] = node
}

func (mts *MockTreeStore)putStaleNodeIndex(index StaleNodeIndex)  {
	if res := mts.staleNode.Add(index); res == false {
		panic("Duplicated retire log.")
	}
}

func (mts *MockTreeStore)numNodes() int {
	return len(mts.data)
}

// purge stale nodes
func (mts *MockTreeStore)purgeStaleNodes(leastReadableVersion Version)  {
	var toPurge []StaleNodeIndex
	for item := range mts.staleNode.Iter() {
		staleInx := item.(StaleNodeIndex)
		if staleInx.StaleSinceVersion <= leastReadableVersion {
			toPurge = append(toPurge, staleInx)
		}
	}
	for _, p := range toPurge {
		delete(mts.data, p.NodeK)
		mts.staleNode.Remove(p)
	}
}
