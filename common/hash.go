package common

import "crypto/sha256"

const HashLength int = 32

type HashValue [HashLength]byte

type CryptoHash interface {
	Hash() HashValue
}

type SparseMerkleInternalNode struct {
	LeftNode HashValue
	RightNode HashValue
}

type SparseMerkleLeafNode struct {
	key HashValue
	valueHash HashValue
}

func (h HashValue)Bytes() []byte {return h[:]}

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (h *HashValue) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}
	copy(h[HashLength-len(b):], b)
}

func (s SparseMerkleInternalNode)Hash() HashValue {
	var res HashValue
	hashBytes := append(s.LeftNode.Bytes(), s.RightNode.Bytes()...)
	hashes := sha256.Sum256(hashBytes)
	res.SetBytes(hashes[:])
	return res
}

func (s SparseMerkleLeafNode)Hash() HashValue {
	var res HashValue
	hashBytes := append(s.key.Bytes(), s.valueHash.Bytes()...)
	hashes := sha256.Sum256(hashBytes)
	res.SetBytes(hashes[:])
	return res
}

func BytesToHash(b []byte) HashValue {
	var h HashValue
	h.SetBytes(b)
	return h
}

func CountOnes(u uint16) int{
	return 7
}

func TrailingZeros(u uint16) int{
	return 7
}
