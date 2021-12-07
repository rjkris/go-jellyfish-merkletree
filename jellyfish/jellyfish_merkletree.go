package jellyfish

import (
	"crypto/sha256"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

type TreeReader interface {
	getNode(nodeKey NodeKey) (Node, error)
	getRightMostLeaf() LeafNode
}

type JfValue interface {  // TODO: 接口定义
    // New([]uint8) JfValue
    GetValue() []byte
}

type ValueT struct {
	Value []byte
}

type TreeWriter interface {
	WriteTreeUpdateBatch(batch TreeUpdateBatch) error
}

type NodeBatch map[NodeKey]Node  // NodeKey类型不能比较
type StaleNodeIndexBatch mapset.Set  // StaleNodeIndex

type NodeStats struct {
	NewNodes uint
	NewLeaves uint
	StaleNodes uint
	StaleLeaves uint
}

/// Indicates a node becomes stale since `stale_since_version`.
type StaleNodeIndex struct {
	StaleSinceVersion Version
	NodeK NodeKey
}

type TreeUpdateBatch struct {
	NodeBch NodeBatch
	StaleNodeIndexBch StaleNodeIndexBatch
	NodeStatsList []NodeStats
}

type JfMerkleTree struct {
	Reader TreeReader  // TODO:  reader接口化
	Value interface{}
}

type ValueSetItem struct {
	HashK common.HashValue
	Value JfValue
}

func (v ValueT) GetValue() []byte {
	return v.Value
}

func (v ValueT)Hash() common.HashValue {
	valueHash := sha256.Sum256(v.GetValue())
	return valueHash
}

func (jf *JfMerkleTree)treeGetValue(key common.HashValue, version Version) JfValue {
	res, _ := jf.getWithProof(key, version)
	return res
}

func (jf *JfMerkleTree)PutValueSet(valueSet []ValueSetItem, version Version) (common.HashValue, TreeUpdateBatch) {
	rootHashList, treeUpdateBatch := jf.PutValueSets([][]ValueSetItem{valueSet}, version)
	if len(rootHashList) != 1 {
		panic("root_hashes must consist of a single Value.")
	}
	return rootHashList[0], treeUpdateBatch
}

func (jf *JfMerkleTree)PutValueSets(valueSets [][]ValueSetItem, firstVersion Version) ([]common.HashValue, TreeUpdateBatch) {
	treeCache := TreeCache{}.new(jf.Reader, firstVersion)

	for i, valueSet := range valueSets {
		if len(valueSet) == 0 {
			panic("Transactions that output empty write set should not be included.")
		}
		version := firstVersion+Version(i)
		for _, item := range valueSet {
			jf.put(item.HashK, item.Value, version, treeCache)
		}
		treeCache.freeze()
	}
	return treeCache.into()
}

func (jf *JfMerkleTree)put(key common.HashValue, value JfValue, version Version, treeCa *TreeCache)  {
	nibblePath := NibblePath{}.new(key.Bytes())
	// Get the root node. If this is the first operation, it would get the root node from the
	// underlying Db. Otherwise it most likely would come from `cache`.
	rootNodeKey := treeCa.rootNodeKey
	nibbleIter := nibblePath.nibbles()
	//cloneRootNodeKey := rootNodeKey
	newRootNodeKey, _ := jf.insertAt(rootNodeKey, version, nibbleIter, value, treeCa)
	// fmt.Printf("after insertAt: %+v \n", treeCa)
	treeCa.rootNodeKey = newRootNodeKey
	// fmt.Printf("after update treeCa: %+v \n", treeCa)

}

/// Helper function for recursive insertion into the subtree that starts from the current
/// [`NodeKey`](node_type/struct.NodeKey.html). Returns the newly inserted node.
/// It is safe to use recursion here because the max depth is limited by the key length which
/// for this tree is the length of the hash of account addresses.
func (jf *JfMerkleTree)insertAt(nodeK NodeKey, version Version, nibbleIter *NibbleIterator, value JfValue, treeCa *TreeCache) (NodeKey, Node) {
	// fmt.Println("debuggggg insertAt")
	node := treeCa.getNode(nodeK)
	switch inode:= node.(type) {
	case InternalNode:
		return jf.insertAtInternalNode(nodeK, inode, version, nibbleIter, value, treeCa)
	case LeafNode:
		return jf.insertAtLeafNode(nodeK, inode, version, nibbleIter, value, treeCa)
	case NoneNode:
		// fmt.Println("nonenode is running")
		if nodeK.Np.NumNibbles != 0 {
			panic("null node exists for non-root node with nodeKey")
		}
		if nodeK.Vs == version {
			treeCa.deleteNode(nodeK, false)
		}
		// fmt.Printf("after delete node: %+v \n", treeCa)
		return jf.createLeafNode(NodeKey{}.newEmptyPath(version), nibbleIter, value, treeCa)
	default:
		panic("unknown node type")
		return NodeKey{}, nil
	}
}

func (jf *JfMerkleTree)insertAtInternalNode(nodeK NodeKey, node InternalNode, version Version, nibbleIter *NibbleIterator, value JfValue, treeCa *TreeCache) (NodeKey, Node) {
	// first delete and put in the end
	treeCa.deleteNode(nodeK, false)
	childIndex := nibbleIter.next()
	if childIndex == nil{
		panic("ran out of nibbles")
	}
	var newChildNode Node
	nodeChild := node.child(childIndex.(Nibble))
	if (nodeChild == Child{}) {
		newChildNodeKey := nodeK.genChildNodeKey(version, childIndex.(Nibble))
		_, newChildNode = jf.createLeafNode(newChildNodeKey, nibbleIter, value, treeCa)
	} else {
		childNodeKey := nodeK.genChildNodeKey(nodeChild.Vs, childIndex.(Nibble))
		_, newChildNode = jf.insertAt(childNodeKey, version, nibbleIter, value, treeCa)
	}
	// Reuse the current `InternalNode` in memory to create a New internal node.
	children := ChildrenClone(node.Children) // can't Children := common.childrenx
	children[childIndex.(Nibble)] = Child{
		Hash:   newChildNode.hash(),
		Vs:     version,
		IsLeaf: newChildNode.isLeaf(),
	}
	newInternalNode := InternalNode{}.new(children)
	nodeK.Vs = version
	err := treeCa.putNode(nodeK, newInternalNode)
    if err != nil {
    	// fmt.Println(err)
	}
	return nodeK, newInternalNode
}

func (jf *JfMerkleTree)insertAtLeafNode(nodeK NodeKey, existingLeafNode LeafNode, version Version, nibbleIter *NibbleIterator, value JfValue, treeCa *TreeCache) (NodeKey, Node) {
	// fmt.Println("debugggggggg insertAtLeafNode")
	treeCa.deleteNode(nodeK, true)
	// fmt.Println("deletenode finished")
	// 1. Make sure that the existing leaf nibble_path has the same prefix as the already
	// visited part of the nibble iter of the incoming key and advances the existing leaf
	// nibble iterator by the length of that prefix.
	visitedNibbleIter := nibbleIter.visitedNibbles()
	existingLeafNibblePath := NibblePath{}.new(existingLeafNode.AccountKey.Bytes())  // current leafNode
	existingLeafNibbleIter := existingLeafNibblePath.nibbles()
	// fmt.Println("debuggggggggg skip")
	SkipCommonPrefix(visitedNibbleIter, existingLeafNibbleIter)
	// fmt.Println("debuggggggggg skip2")
	if visitedNibbleIter.isFinished() == false {
		panic("Leaf nodes failed to share the same visited nibbles before index " + string(existingLeafNibbleIter.visitedNibbles().numNibbles()))
	}

	// 2. Determine the extra part of the common prefix that extends from the position where
	// step 1 ends between this leaf node and the incoming key.
	existingLeafNibbleIterBelowInternal := existingLeafNibbleIter.remainingNibbles()
	numCommonNibblesBelowInternal := SkipCommonPrefix(nibbleIter, existingLeafNibbleIterBelowInternal)
	// fmt.Println("debuggggggggg skip3")
	commonNibblePath := nibbleIter.visitedNibbles().getPartNibblePath() // get common nibblePath
	// 2.1. Both are finished. That means the incoming key already exists in the tree and we
	// just need to update its Value.
	if nibbleIter.isFinished() {
		if existingLeafNibbleIterBelowInternal.isFinished() == false {
			panic("insert leafNode error")
		}
		nodeK.Vs = version
		// fmt.Printf("insertatleafnode 11111111111")
		return jf.createLeafNode(nodeK, nibbleIter, value, treeCa)
	}
	// 2.2. both are unfinished(They have keys with same length so it's impossible to have one
	// finished and the other not). This means the incoming key forks at some point between the
	// position where step 1 ends and the last nibble, inclusive. Then create a seris of
	// internal nodes the number of which equals to the length of the extra part of the
	// common prefix in step 2, a New leaf node for the incoming key, and update the
	// [`NodeKey`] of existing leaf node. We create New internal nodes in a bottom-up
	// order.
	existingLeafIndex := existingLeafNibbleIterBelowInternal.next()
	if existingLeafIndex == nil {
		panic("ran out of nibbles")
	}
	newLeafIndex := nibbleIter.next()
	if newLeafIndex == nil {
		panic("ran out of nibbles")
	}
	children := Children{}
	children[existingLeafIndex.(Nibble)] = Child{
		Hash:   existingLeafNode.hash(),
		Vs:     version,
		IsLeaf: true,
	}
	nodeK = NodeKey{
		Vs: version,
		Np: commonNibblePath,
	}
	err := treeCa.putNode(nodeK.genChildNodeKey(version, existingLeafIndex.(Nibble)), existingLeafNode)
	if err != nil {
		_ = fmt.Errorf("put node error: %w", err)
	}
	// fmt.Println("debuggggggggg before createleafnode")
	_, newLeafNode := jf.createLeafNode(nodeK.genChildNodeKey(version, newLeafIndex.(Nibble)), nibbleIter, value, treeCa)
	children[newLeafIndex.(Nibble)] = Child{
		Hash:   newLeafNode.hash(),
		Vs:     version,
		IsLeaf: true,
	}

	internalNode := InternalNode{}.new(children)  // New internalNode which include
	nextInternalNode := internalNode
	err = treeCa.putNode(nodeK, internalNode)
	if err != nil {
		panic(fmt.Sprintf("put node error: %s", err))
	}
	for i :=0; i<int(numCommonNibblesBelowInternal); i++ {
		// fmt.Printf("debug forrrrrrrrrrrrrrr")
		nibble, _ := commonNibblePath.pop()
		nodeK = NodeKey{
			Vs: version,
			Np: commonNibblePath,
		}
		children := Children{}
		children[nibble] = Child{
			Hash:   nextInternalNode.hash(),
			Vs:     version,
			IsLeaf: false,
		}
		internalNode := internalNode.new(children)
		nextInternalNode = internalNode
		err := treeCa.putNode(nodeK, internalNode)
		if err != nil {
			panic(fmt.Sprintf("put node error: %s", err))
		}
	}
    return nodeK, nextInternalNode
}

func (jf *JfMerkleTree)createLeafNode(nodeK NodeKey, nibbleIter *NibbleIterator, value JfValue, treeCa *TreeCache) (NodeKey, Node) {
	var newkey common.HashValue
	// Get the underlying bytes of nibble_iter which must be a key, i.e., hashed account address
	// with `HashValue::LENGTH` bytes.
	newkey.SetBytes(nibbleIter.getNibblePath().Bytes)
	newLeafNode := LeafNode{}.newLeaf(newkey, value)
	err := treeCa.putNode(nodeK, newLeafNode)
	if err != nil {
		panic(fmt.Sprint(err))
	}
	return nodeK, newLeafNode
}

func (jf *JfMerkleTree)getWithProof(key common.HashValue, version Version) (JfValue, SparseMerkleProof) {
	nextNodeKey := NodeKey{}.newEmptyPath(version)
	var siblings []common.HashValue
	nibblePath := NibblePath{}.new(key.Bytes())
	nibbleIter := nibblePath.nibbles()
	for nibbleDepth :=0; nibbleDepth <=common.RootNibbleHeight; nibbleDepth++ {
		// fmt.Printf("current nextNode: %+v \n", nextNodeKey)
		// fmt.Println("debugggggggggggggggggggggg")
		nextNode, _ := jf.Reader.getNode(nextNodeKey)
		//if err != nil {
		//	panic(err)
		//}
		switch node := nextNode.(type) {
		case InternalNode:
			queriedChildIndex := nibbleIter.next()
			if queriedChildIndex == nil {
				panic("ran out of nibbles")
			}
			childNodeKey, siblingsInInternal := node.getChildWithSiblings(nextNodeKey, queriedChildIndex.(Nibble))
			// fmt.Printf("childnodekey: %+v \n", childNodeKey)
			siblings = append(siblings, siblingsInInternal...)
			if childNodeKey == nil  {
				// fmt.Println("proof111111111")
				return nil, SparseMerkleProof{
					leaf:     common.SparseMerkleLeafNode{},
					siblings: common.Reverse(siblings),
				}
			} else {
				nextNodeKey = childNodeKey.(NodeKey)
			}
		case LeafNode:
			if node.AccountKey == key {
				// fmt.Println("proof2222222222")
				return node.Value, SparseMerkleProof{common.SparseMerkleLeafNode{node.AccountKey, node.ValueHash}, common.Reverse(siblings)}
			} else {
				// fmt.Println("proof2222222222")
				return nil, SparseMerkleProof{common.SparseMerkleLeafNode{node.AccountKey, node.ValueHash}, common.Reverse(siblings)}
			}
		case NoneNode:
			// fmt.Printf("getnode nonenode")
			if nibbleDepth == 0 {
				// fmt.Println("proof3333333333")
				return nil, SparseMerkleProof{
					leaf:     common.SparseMerkleLeafNode{},
					siblings: []common.HashValue{},
				}
			} else {
				// fmt.Printf("non-root null node exists with node key: %v", nextNodeKey)
			}
		}
	}
	// fmt.Println("Jellyfish Merkle tree has cyclic graph inside.")
	return nil, SparseMerkleProof{}
}


func ChildrenClone(children Children) Children {
	newChildren := make(map[Nibble]Child)
	for k, v := range children {
		newChildren[k] = v
	}
	return newChildren
}
//TODO: add getRangeProof
//func (jf *JfMerkleTree)getRangeProof(rightmostKeyToProve common.HashValue, version Version) SparseMerkleProof {
//	account, proof := jf.getWithProof(rightmostKeyToProve, version)
//	if account == nil {
//		panic("rightmostKeyToProve must exist.")
//	}
//	siblings := proof.siblings
//}