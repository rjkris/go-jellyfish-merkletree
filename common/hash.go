package common

import (
	"crypto/sha256"
	mapset "github.com/deckarep/golang-set"
	"math/rand"
	"time"
)

const HashLength int = 32
const RootNibbleHeight int = HashLength*2


type HashValue [HashLength]byte

type CryptoHash interface {
	Hash() HashValue
}

type SparseMerkleInternalNode struct {
	LeftNode HashValue
	RightNode HashValue
}

type SparseMerkleLeafNode struct {
	Key HashValue
	ValueHash HashValue
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

func (h HashValue)Random() HashValue {
	res := make([]byte, HashLength)
	rand.Seed(time.Now().UnixNano())
	_, err := rand.Read(res)
	if err != nil {
		println(err)
	}
	h.SetBytes(res)
	return h
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
	hashBytes := append(s.Key.Bytes(), s.ValueHash.Bytes()...)
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
	var count int
	for i:=0; i<16; i++ {
		count += int(u) & 1
		u >>= 1
	}
	return count
}

func TrailingZeros(u uint16) int{
	var count int
	for i:=0; i<16; i++ {
		if u & 1 == 1 {
			break
		} else {
			u >>= 1
			count += 1
		}
	}
	return count
}

// TODO: DEBUG
func LeadingZeros(u uint16) int{
	var count int
	var tmp = uint16(1 << 15)
	for i:=0; i<16; i++ {
		if u & tmp == 1 {
			break
		} else {
			u <<= 1
			count += 1
		}
	}
	return count
}

func golangSetLen(s mapset.Set) {

}
