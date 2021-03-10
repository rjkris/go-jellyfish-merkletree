package jellyfish

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"go-jellyfish-merkletree/common"
)

type FrozenTreeCache struct {
	nodeCache map[NodeKey]Node
	staleNodeIndexCache mapset.Set  // StaleNodeIndex
	nodeStatsList []NodeStats
	rootHashesList []common.HashValue
}

type TreeCache struct {
	rootNodeKey *NodeKey
	nextVersion Version
	nodeCache map[NodeKey]Node
	numNewLeaves uint
	StaleNodeIndexCache mapset.Set  // NodeKey
	numStaleLeaves uint
	frozenCache FrozenTreeCache
	reader TreeReader
}

func (tc TreeCache)new(reader TreeReader, nextVersion Version) *TreeCache {
	nodeCache := map[NodeKey]Node{}
	frozenNodeCache := FrozenTreeCache{
		nodeCache:           map[NodeKey]Node{},
		staleNodeIndexCache: mapset.NewSet(),
	}
	var rootNodeKey *NodeKey
	// extreme case
	if nextVersion == 0 {
		preGenesisRootKey := NodeKey{}.newEmptyPath(PreGenesisVersion)
		preGenesisRoot := reader.getNode(preGenesisRootKey)
		_, ok := preGenesisRoot.(Node)
		if ok {
			rootNodeKey = preGenesisRootKey
		}else {
			fmt.Println("---------------------")
			genesisRootKey := NodeKey{}.newEmptyPath(0)
			nodeCache[*genesisRootKey] = NoneNode{}
			rootNodeKey = genesisRootKey
			fmt.Printf("genesisRootKey: %p \n", rootNodeKey)
		}
	}else {
		rootNodeKey = NodeKey{}.newEmptyPath(nextVersion-1)
	}
	return &TreeCache{
		rootNodeKey:         rootNodeKey,
		nextVersion:         nextVersion,
		nodeCache:           nodeCache,
		numNewLeaves:        0,
		StaleNodeIndexCache: mapset.NewSet(),
		numStaleLeaves:      0,
		frozenCache:         frozenNodeCache,
		reader:              reader,
	}
}

func (tc *TreeCache)getNode(nodeK *NodeKey) interface{} {
	fmt.Printf("nodeKey value: %p \n", nodeK)
	fmt.Printf("treeCache value: %+v \n", tc)
	if node, ok := tc.nodeCache[*nodeK]; ok {
		fmt.Println("111111111111")
		return node
	} else if node, ok := tc.frozenCache.nodeCache[*nodeK]; ok {
		fmt.Println("222222222222222")
		return node
	} else {
		node := tc.reader.getNode(nodeK)
		fmt.Println("333333333333333")
		return node
	}
}

/// Puts the node with given hash as key into node_cache.
func (tc *TreeCache)putNode(nodeK *NodeKey, newNode Node) error {
	_, ok := tc.nodeCache[*nodeK]
	if ok {
		return fmt.Errorf("node with %v already exists in NodeBatch", nodeK)
	} else {
		if newNode.isLeaf(){
			tc.numNewLeaves += 1
		}
		tc.nodeCache[*nodeK] = newNode
	}
	return nil
}

func (tc *TreeCache)deleteNode(oldNodeKey *NodeKey, isLeaf bool) {
	// If node cache doesn't have this node, it means the node is in the previous version of
	// the tree on the disk.
	if _, ok := tc.nodeCache[*oldNodeKey]; ok == false {
		fmt.Println("delenode false")
		cloneOldNodeKey := *oldNodeKey
		isNewEntry := tc.StaleNodeIndexCache.Add(cloneOldNodeKey)  // TODO: CLONE
		if isNewEntry == false {
			panic("Node gets stale twice unexpectedly")
		}
		if isLeaf == true {
			tc.numStaleLeaves += 1
		}
		return
	}
	delete(tc.nodeCache, *oldNodeKey)
	if isLeaf == true {
		tc.numStaleLeaves -= 1
	}
}

// Freezes all the contents in cache to be immutable and clear `node_cache`.
func (tc *TreeCache)freeze()  {
	rootNode := tc.getNode(tc.rootNodeKey)
	fmt.Printf("rootNode type: %T", rootNode)
	//if rootNode ==  NoneNode{} {
	//	panic("Root node must exit")
	//}
	rootHash := rootNode.(Node).hash()
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
	fmt.Printf("current frozencache: %+v \n", tc.frozenCache)
	tc.nodeCache = map[NodeKey]Node{}
	staleSinceVersion := tc.nextVersion
	for item := range tc.StaleNodeIndexCache.Iter() {
		fmt.Println("iiiiiiiiiiiiii am iterrrrrrrrrr")
		nodeK, _ := item.(NodeKey)
		tc.frozenCache.staleNodeIndexCache.Add(StaleNodeIndex{
			StaleSinceVersion: staleSinceVersion,
			NodeK:             nodeK,
		})
	}
	//tc.StaleNodeIndexCache = mapset.NewSet()
	tc.numStaleLeaves = 0
	tc.numNewLeaves = 0
	tc.nextVersion += 1
}

func (tc *TreeCache)into() ([]common.HashValue, TreeUpdateBatch) {
	fmt.Printf("into value: %v \n", tc.frozenCache.rootHashesList)
	return tc.frozenCache.rootHashesList, TreeUpdateBatch{
		NodeBch:           tc.frozenCache.nodeCache,
		StaleNodeIndexBch: tc.frozenCache.staleNodeIndexCache,
		NodeStatsList:     tc.frozenCache.nodeStatsList,
	}
}


