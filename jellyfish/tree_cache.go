package jellyfish

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"go-jellyfish-merkletree/common"
)

type FrozenTreeCache struct {
	nodeCache map[*NodeKey]Node
	staleNodeIndexCache mapset.Set  // StaleNodeIndex
	nodeStatsList []NodeStats
	rootHashesList []common.HashValue
}

type TreeCache struct {
	rootNodeKey NodeKey
	nextVersion Version
	nodeCache map[*NodeKey]Node
	numNewLeaves uint
	StaleNodeIndexCache mapset.Set  // NodeKey
	numStaleLeaves uint
	frozenCache FrozenTreeCache
	reader TreeReader
}

func (tc *TreeCache)new(reader TreeReader, nextVersion Version) TreeCache {
	nodeCache := map[*NodeKey]Node{}
	var rootNodeKey NodeKey
	// extreme case
	if nextVersion == 0 {
		preGenesisRootKey := NodeKey{}.newEmptyPath(PreGenesisVersion)
		preGenesisRoot, _ := reader.getNode(&preGenesisRootKey)
		_, ok := preGenesisRoot.(Node)
		if ok {
			rootNodeKey = preGenesisRootKey
		}else {
			genesisRootKey := NodeKey{}.newEmptyPath(0)
			nodeCache[&genesisRootKey] = nil  // TODO: 使用拷贝genesisRootKey
			rootNodeKey = genesisRootKey
		}
	}else {
		rootNodeKey = NodeKey{}.newEmptyPath(nextVersion-1)
	}
	return TreeCache{
		rootNodeKey:         rootNodeKey,
		nextVersion:         nextVersion,
		nodeCache:           nodeCache,
		numNewLeaves:        0,
		StaleNodeIndexCache: mapset.NewSet(),
		numStaleLeaves:      0,
		frozenCache:         FrozenTreeCache{},
		reader:              reader,
	}
}

func (tc *TreeCache)getNode(nodeK *NodeKey) Node {
	if node, ok := tc.nodeCache[nodeK]; ok {
		return node
	} else if node, ok := tc.frozenCache.nodeCache[nodeK]; ok {
		return node
	} else {
		node, _ := tc.reader.getNode(nodeK)
		return node.(Node)
	}
}

/// Puts the node with given hash as key into node_cache.
func (tc *TreeCache)putNode(nodeK NodeKey, newNode Node) error {
	_, ok := tc.nodeCache[&nodeK]
	if ok {
		return fmt.Errorf("node with %v already exists in NodeBatch", nodeK)
	} else {
		if newNode.isLeaf(){
			tc.numNewLeaves += 1
		}
		tc.nodeCache[&nodeK] = newNode
	}
	return nil
}

func (tc *TreeCache)deleteNode(oldNodeKey *NodeKey, isLeaf bool) {
	if _, ok := tc.nodeCache[oldNodeKey]; ok == false {
		isNewEntry := tc.StaleNodeIndexCache.Add(oldNodeKey)  // TODO: CLONE
		if isNewEntry == false {
			panic("Node gets stale twice unexpectedly")
		}
		if isLeaf == true {
			tc.numStaleLeaves += 1
		}
	} else if isLeaf == true {
		tc.numStaleLeaves -= 1
	}
}

// Freezes all the contents in cache to be immutable and clear `node_cache`.
func (tc *TreeCache)freeze()  {
	rootNode := tc.getNode(&tc.rootNodeKey)
	if rootNode == nil {
		panic("Root node must exit")
	}
	rootHash := rootNode.hash()
	tc.frozenCache.rootHashesList = append(tc.frozenCache.rootHashesList, rootHash)
    nodeStats := NodeStats{
		NewNodes:    uint(len(tc.nodeCache)),
		NewLeaves:   tc.numNewLeaves,
		StaleNodes:  uint(tc.StaleNodeIndexCache.Cardinality()),
		StaleLeaves: tc.numStaleLeaves,
	}
	tc.frozenCache.nodeStatsList = append(tc.frozenCache.nodeStatsList, nodeStats)
	for k, v := range tc.nodeCache {
		tc.frozenCache.nodeCache[k] = v
	}
	tc.nodeCache = map[*NodeKey]Node{}
	staleSinceVersion := tc.nextVersion
	for item := range tc.StaleNodeIndexCache.Iterator().C {
		nodeK, _ := item.(NodeKey)
		tc.frozenCache.staleNodeIndexCache.Add(StaleNodeIndex{
			StaleSinceVersion: staleSinceVersion,
			NodeK:             nodeK,
		})
	}
	tc.StaleNodeIndexCache = mapset.NewSet()
	tc.numStaleLeaves = 0
	tc.numNewLeaves = 0
	tc.nextVersion += 1
}

func (tc *TreeCache)into() ([]common.HashValue, TreeUpdateBatch) {
	return tc.frozenCache.rootHashesList, TreeUpdateBatch{
		NodeBch:           tc.frozenCache.nodeCache,
		StaleNodeIndexBch: tc.frozenCache.staleNodeIndexCache,
		NodeStatsList:     tc.frozenCache.nodeStatsList,
	}
}


