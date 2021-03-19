package jellyfish

import (
	"crypto/sha256"
	"github.com/rjkris/go-jellyfish-merkletree/common"
)

type Version uint64

const PreGenesisVersion Version = 1<<64-1

type NodeKey struct {
	Vs Version
	Np NibblePath
}

type Child struct {
	Hash   common.HashValue
	Vs     Version
	IsLeaf bool
}

// [`Children`] is just a collection of Children belonging to a [`InternalNode`], indexed from 0 to
// 15, inclusive.
type Children map[Nibble]Child

type InternalNode struct {
	Children Children
}

type LeafNode struct {
	AccountKey common.HashValue
	ValueHash common.HashValue
	Value ValueT
}

type NoneNode struct {

}

func createLiteralHash(word string) common.HashValue{
	var res common.HashValue
	res = common.BytesToHash([]byte(word))
	return res
}

func (nk NodeKey)newEmptyPath(version Version) NodeKey {
	return NodeKey{version, *NibblePath{}.new([]uint8{})}
}
// Generates a child node key based on this node key.
func (nk NodeKey) genChildNodeKey(v Version, n Nibble) NodeKey {
	nodeNibblePath := nk.Np
	(&nodeNibblePath).push(n)
	return NodeKey{v, nodeNibblePath}
}

// Generates parent node key at the same Vs based on this node key.
func (nk NodeKey) genParentNodeKey() NodeKey {
	nodeNibblePath := nk.Np
	if nodeNibblePath.NumNibbles == 1 {
		panic("Current node key is root")
	}
	_, _ = nodeNibblePath.pop()
	return NodeKey{nk.Vs, nodeNibblePath}
}

func (nk NodeKey) Encode() {

}

func (nk NodeKey) Decode() {

}

func (internal InternalNode) new(children Children) InternalNode {
	if len(children) == 0 {
		panic("Children is empty")
	}
	if len(children) == 1 {
		for _, value := range children {
			if value.IsLeaf {
				panic("Child can't be a leaf node")
			}
		}
	}
	return InternalNode{children}
}


func (internal InternalNode)isLeaf() bool {
	return false
}

func (internal InternalNode)newLeaf(accountK common.HashValue, value JfValue) Node {
	return nil
}

func (internal InternalNode) hash() common.HashValue{
	existenceBp, leafBp := internal.generateBitmaps()
	return internal.merkleHash(0, 16, existenceBp, leafBp)
}

func (internal *InternalNode) serialize() {

}

func (internal *InternalNode) deserialize() {

}

// Gets the `n`-th child.
func (internal *InternalNode) child(n Nibble) Child {
	return internal.Children[n]
}

// Generates `existence_bitmap` and `leaf_bitmap` as a pair of `u16`s: child at index `i`
// exists if `existence_bitmap[i]` is set; child at index `i` is leaf node if
// `leaf_bitmap[i]` is set.
func (internal *InternalNode) generateBitmaps() (uint16, uint16) {
	var existenceBitmap uint16
	var leafBitmap uint16
	for nibble, child := range internal.Children {
		existenceBitmap |= 1 << nibble
		if child.IsLeaf {
			leafBitmap |= 1 << nibble
		}
	}
	if existenceBitmap|leafBitmap != existenceBitmap {
		panic("GenerateBitmaps error")
	}
	return existenceBitmap, leafBitmap
}

// Given a range [start, start + width), returns the sub-bitmap of that range.
// height 0 with width 1
// TODO: assert
func (internal *InternalNode) rangeBitmaps(start uint8, width uint8, existenceBitmap uint16, leafBitmap uint16) (uint16, uint16) {
	if start >= 16 || width > 16 {
		panic("Start out of range")
	}
	var mask uint16
	if width == 16 {
		mask = 0xffff
	} else {
		mask = (1 << width) - 1
	}
	mask <<= start
	return existenceBitmap & mask, leafBitmap & mask
}

// TODO: understand
func (internal *InternalNode) merkleHash(start uint8, width uint8, existenceBitmap uint16, leafBitmap uint16) common.HashValue {
	var res common.HashValue
	rangeExistenceBitmap, rangeLeafBitmap := internal.rangeBitmaps(start, width, existenceBitmap, leafBitmap)
	if rangeExistenceBitmap == 0 {
		res = common.BytesToHash([]byte("SPARSE_MERKLE_PLACEHOLDER_HASH"))
		// Only 1 leaf child under this subtree or reach the lowest level??
	}else if common.CountOnes(rangeExistenceBitmap) == 1 && (rangeLeafBitmap != 0 || width == 1) {
		onlyChildIndex := Nibble(common.TrailingZeros(rangeExistenceBitmap))
		res = internal.child(onlyChildIndex).Hash
	}else {
		leftChild := internal.merkleHash(start, width/2, existenceBitmap, leafBitmap)
		rightChild := internal.merkleHash(start+width/2, width/2, existenceBitmap, leafBitmap)
		res = common.SparseMerkleInternalNode{LeftNode: leftChild, RightNode: rightChild}.Hash()
	}
	return res
}

/// Gets the child and its corresponding siblings that are necessary to generate the proof for
/// the `n`-th child. If it is an existence proof, the returned child must be the `n`-th
/// child; otherwise, the returned child may be another child. See inline explanation for
/// details. When calling this function with n = 11 (node `b` in the following graph), the
/// range at each level is illustrated as a pair of square brackets:
///
/// ```text
///     4      [f   e   d   c   b   a   9   8   7   6   5   4   3   2   1   0] -> root level
///            ---------------------------------------------------------------
///     3      [f   e   d   c   b   a   9   8] [7   6   5   4   3   2   1   0] width = 8
///                                  chs <--┘                        shs <--┘
///     2      [f   e   d   c] [b   a   9   8] [7   6   5   4] [3   2   1   0] width = 4
///                  shs <--┘               └--> chs
///     1      [f   e] [d   c] [b   a] [9   8] [7   6] [5   4] [3   2] [1   0] width = 2
///                          chs <--┘       └--> shs
///     0      [f] [e] [d] [c] [b] [a] [9] [8] [7] [6] [5] [4] [3] [2] [1] [0] width = 1
///     ^                chs <--┘   └--> shs
///     |   MSB|<---------------------- uint 16 ---------------------------->|LSB
///  height    chs: `child_half_start`         shs: `sibling_half_start`
/// ```
// []common.HashValue len: 4
func (internal *InternalNode) getChildWithSiblings(nodeKey NodeKey, n Nibble) (interface{}, []common.HashValue) {
	var siblings []common.HashValue
	existenceBitmap, leafBitmap := internal.generateBitmaps()
	for h:=uint8(3); h>=0; h-- {
		width := uint8(1 << h)
		childHalfStart, siblingHalfStart := GetChildAndSiblingHalfStart(n, h)
		siblings = append(siblings, internal.merkleHash(siblingHalfStart, width, existenceBitmap, leafBitmap))
	    rangeExistenceBitmap, rangeLeafBitmap := internal.rangeBitmaps(childHalfStart, width, existenceBitmap, leafBitmap)
	    if rangeExistenceBitmap == 0{
	    	return nil, siblings
		}else if common.CountOnes(rangeExistenceBitmap) == 1 && (common.CountOnes(rangeLeafBitmap) == 1 || width == 1){
			onlyChildIndex := uint8(common.TrailingZeros(rangeExistenceBitmap))
			onlyChildVersion := internal.child(onlyChildIndex).Vs
			return nodeKey.genChildNodeKey(onlyChildVersion, onlyChildIndex), siblings
		}
	}
	return nil, nil // unreached
}

func GetChildAndSiblingHalfStart(n Nibble, height uint8) (uint8, uint8) {
	childHalfStart := (0xff << height) & n
	siblingHalfStart := childHalfStart ^ (1 << height)
	return childHalfStart, siblingHalfStart
}

func (lf *LeafNode)new(accountKey common.HashValue, value JfValue) LeafNode {
	valueHash := sha256.Sum256(value.getValue())
	return LeafNode{accountKey, valueHash, value.(ValueT)}
}

func (lf LeafNode)hash() common.HashValue {
	return common.SparseMerkleLeafNode{lf.AccountKey, lf.ValueHash}.Hash()
}

func (lf LeafNode)newLeaf(accountk common.HashValue, value JfValue) Node {
	return lf.new(accountk, value)
}
func (lf LeafNode)isLeaf() bool {
	return true
}

func (n NoneNode)hash() common.HashValue {
	return common.BytesToHash([]byte("SPARSE_MERKLE_PLACEHOLDER_HASH"))
}
func (n NoneNode)newLeaf(key common.HashValue, value JfValue) Node {
	return nil
}
func (n NoneNode)isLeaf() bool {
	return false
}

// TODO 节点接口抽象 internal || leaf || nil
type Node interface {
	// newNone() Node
	// newInternal(Children) Node
	newLeaf(common.HashValue, JfValue) Node
	isLeaf() bool
	// encode() []uint8
	hash() common.HashValue
	// decode(*[]uint8) Node
}






