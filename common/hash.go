package common

import (
	"crypto/sha256"
	"math/rand"
	"time"
)

const HashLength int = 32
const RootNibbleHeight int = HashLength*2
const LengthInBits int = HashLength*8

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
func (h *HashValue) SetBytes(b [HashLength]byte) {
	//if len(b) > len(h) {
	//	b = b[len(b)-HashLength:]
	//}
	//copy(h[HashLength-len(b):], b)
	for i, v := range b {
		h[i] = v
	}
}

func (h HashValue)Random() HashValue {
	res := make([]byte, HashLength)
	var resArray [HashLength]byte
	rand.Seed(time.Now().UnixNano())
	_, err := rand.Read(res)
	if err != nil {
		println(err)
	}
	for i, v := range res {
		resArray[i] = v
	}
	h.SetBytes(resArray)
	return h
}

func (h HashValue)Bytes2Bits() []int {
	res := make([]int, HashLength*32)
	for i, v := range h {
		for j:=0; j<8; j++ {
			res[i*8+j] = int(v >> uint(7-j) & 0x01)
		}
	}
	return res
}

//func (h HashValue)commonPrefixBitsLen(other HashValue) uint {
//
//}

func (s SparseMerkleInternalNode)Hash() HashValue {
	var res HashValue
	hashBytes := append(s.LeftNode.Bytes(), s.RightNode.Bytes()...)
	hashes := sha256.Sum256(hashBytes)
	res.SetBytes(hashes)
	return res
}

func (s SparseMerkleLeafNode)Hash() HashValue {
	var res HashValue
	hashBytes := append(s.Key.Bytes(), s.ValueHash.Bytes()...)
	hashes := sha256.Sum256(hashBytes)
	res.SetBytes(hashes)
	return res
}

func BytesToHash(b []byte) HashValue {
	var h HashValue
	var byteArray [HashLength]byte
	for i, v := range b {
		byteArray[i] = v
	}
	h.SetBytes(byteArray)
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

func Reverse(s []HashValue) []HashValue {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
