package jellyfish

import "github.com/rjkris/go-jellyfish-merkletree/common"

type NodeVisitInfo struct {
	nodeK NodeKey
	node InternalNode
	childrenBitmap uint16
	/// This integer always has exactly one 1-bit. The position of the 1-bit (from LSB) indicates
	/// the next child to visit in the iteration process. All the ones on the left have already
	/// been visited. All the chilren on the right (including this one) have not been visited yet.
	nextChildToVisit uint16
}

// TODO: storage engine field
type JfMerkleIterator struct {
	reader interface{}
	vs Version
	parentStack []NodeVisitInfo
	done bool
	value interface{}
}

func (nodeVI *NodeVisitInfo)new(nodeK NodeKey, node InternalNode) NodeVisitInfo {
	childrenBitmap, _ := node.generateBitmaps()
	return NodeVisitInfo{nodeK, node, childrenBitmap, 1 << common.TrailingZeros(childrenBitmap)}
}

func (nodeVI *NodeVisitInfo)newNextChildToVisit(nodeK NodeKey, node InternalNode, nextChildToVisit Nibble) NodeVisitInfo{
	childrenBitmap, _ := node.generateBitmaps()
	var nextChildToVisitU uint16
	nextChildToVisitU = 1 << nextChildToVisit
	for nextChildToVisitU & childrenBitmap == 0{
		nextChildToVisitU <<= 1
	}
	return NodeVisitInfo{nodeK, node, childrenBitmap, nextChildToVisitU}
}

func (nodeVI *NodeVisitInfo)isRightMost() bool {
	if common.LeadingZeros(nodeVI.nextChildToVisit) < common.LeadingZeros(nodeVI.childrenBitmap) {
		panic("isRightMost error")
	}
	return common.LeadingZeros(nodeVI.nextChildToVisit) == common.LeadingZeros(nodeVI.childrenBitmap)
}

func (nodeVI *NodeVisitInfo)advance()  {
	if nodeVI.isRightMost(){
		panic("Advancing past rightMost child.")
	}
	nodeVI.nextChildToVisit <<= 1
	for nodeVI.nextChildToVisit & nodeVI.childrenBitmap == 0 {
		nodeVI.nextChildToVisit <<= 1
	}
}
//
//func (jfIterator *JfMerkleIterator)New(reader interface{}, version Version, startingKey common.HashValue) JfMerkleIterator {
//	var parentStack []NodeVisitInfo
//	var done bool = false
//	currentNodeKey := NodeKey{}.newEmptyPath(version)
//	nibblePath := NibblePath{}.New(startingKey.Bytes())
//	nibbleIter := nibblePath.nibbles()
//	for node := Node{Internal: }
//}

