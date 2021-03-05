package jellyfish

import (
	"fmt"
	"go-jellyfish-merkletree/common"
)

type NibblePath struct {
	NumNibbles int
	Bytes      []uint8
}

type BitIterator struct {
	nibblePath NibblePath
	start int
	end int
}

type NibbleIterator struct {
	nibblePath NibblePath
	start int
	cur int
	end int
}

type interator interface {
	peek() interface{}
	next() interface{}
}

type Nibble = uint8

// Creates a new `NibblePath` from a vector of Bytes assuming each byte has 2 nibbles.
func (np NibblePath) new(bytes []uint8) NibblePath {
	// TODO: Bytes len check
	numNibbles := len(bytes) * 2
	return NibblePath{numNibbles, bytes}
}

// NumNibbles is odd
func (np *NibblePath) newOdd(bytes []uint8) (NibblePath, error) {
	// TODO: Bytes len check
	if bytes[len(bytes)-1]&0x0f != 0 {
		return NibblePath{}, fmt.Errorf("last nibble must be 0")
	}
	numNibbles := len(bytes)*2 - 1
	return NibblePath{numNibbles, bytes}, nil
}

// Adds a nibble to the end of the nibble path
func (np *NibblePath) push(nibble Nibble) {
	// TODO: Bytes len check
	if np.NumNibbles%2 == 0 {
		np.Bytes = append(np.Bytes, nibble<<4)
	} else {
		np.Bytes[np.NumNibbles/2] |= nibble
	}
	np.NumNibbles += 1
}

// Pops a nibble from the end of the nibble path.
func (np *NibblePath) pop() (Nibble, error) {
	var lastNibble Nibble
	if np.NumNibbles <= 0 {
		return lastNibble, fmt.Errorf("nibblePath is empty")
	}
	l := len(np.Bytes)
	if np.NumNibbles%2 == 0 {
		lastNibble = np.Bytes[l-1] & 0x0f
		np.Bytes[l-1] &= 0xf0
	} else {
		lastNibble = np.Bytes[l-1] >> 4
		np.Bytes = np.Bytes[:l-1]
	}
	np.NumNibbles -= 1
	return lastNibble, nil
}

// Returns the last nibble.
func (np *NibblePath) last() (Nibble, error) {
	var lastNibble Nibble
	if np.NumNibbles <= 0 {
		return lastNibble, fmt.Errorf("nibblePath is empty")
	}
	lastByte := np.Bytes[len(np.Bytes)-1]
	if np.NumNibbles%2 == 0 {
		lastNibble = lastByte & 0x0f
	} else {
		lastNibble = lastByte >> 4
	}
	return lastNibble, nil
}

// Get the i-th bit.
func (np *NibblePath) getBit(i int) bool {
	if i/4 >= np.NumNibbles {
		panic("i out of nibblePath range")
	}
	pos := i / 8
	bit := 7 - i%8
	return (np.Bytes[pos] >> bit) != 0
}

// Get the i-th nibble
func (np *NibblePath) getNibble(i int) Nibble {
	if i >= np.NumNibbles {
		panic("i out of nibblePath range")
	}
	if i%2 == 1 {
		return (np.Bytes[i/2]) & 0xf
	} else {
		return (np.Bytes[i/2] >> 4) & 0xf
	}
}

func (np NibblePath) bits() BitIterator{
	if np.NumNibbles > common.RootNibbleHeight {
		panic("out of range")
	}
	return BitIterator{np, 0, np.NumNibbles*4}
}

func (np *NibblePath) nibbles() NibbleIterator{
	if np.NumNibbles > common.RootNibbleHeight {
		panic("out of range")
	}
	return NibbleIterator{}.new(*np,0, np.NumNibbles)
}

func (bIter BitIterator)peek() interface{} {
	if bIter.start < bIter.end {
		return bIter.nibblePath.getBit(bIter.start)
	}else {
		return nil
	}
}

func (bIter BitIterator)next() interface{} {
	if bIter.start < bIter.end {
		res := bIter.nibblePath.getBit(bIter.start)
		bIter.start ++
		return res
	}else {
		return nil
	}
}

func (bIter *BitIterator)nextBack() interface{} {
	if bIter.start < bIter.end {
		res := bIter.nibblePath.getBit(bIter.end)
		bIter.end --
		return res
	}else {
		return nil
	}
}

func (nIter NibbleIterator)next() interface{} {
	if nIter.cur < nIter.end {
		res := nIter.nibblePath.getNibble(nIter.cur)
		nIter.cur ++
		return res
	}else {
		return nil
	}
}

func (nIter NibbleIterator)peek() interface{} {
	if nIter.cur < nIter.end {
		return nIter.nibblePath.getNibble(nIter.cur)
	}else {
		return nil
	}
}


func (nIter NibbleIterator)new(nibblePath NibblePath, start int, end int) NibbleIterator {
	if start > end || start > common.RootNibbleHeight|| end > common.RootNibbleHeight{
		panic("out of range")
	}else {
		return NibbleIterator{nibblePath, start, start, end}
	}
}

// Returns a nibble iterator that iterates all visited nibbles.
func (nIter *NibbleIterator)visitedNibbles() NibbleIterator {
	if nIter.start > nIter.cur || nIter.cur > common.RootNibbleHeight {
		panic("out of range")
	}
	return NibbleIterator{}.new(nIter.nibblePath, nIter.start, nIter.cur)
}

// Returns a nibble iterator that iterates all remaining nibbles.
func (nIter *NibbleIterator)remainingNibbles() NibbleIterator {
	if nIter.cur > nIter.end || nIter.end > common.RootNibbleHeight {
		panic("out of range")
	}
	return NibbleIterator{}.new(nIter.nibblePath, nIter.cur, nIter.end)
}

func (nIter *NibbleIterator)bits() BitIterator {
	if nIter.cur > nIter.end || nIter.end > common.RootNibbleHeight {
		panic("out of range")
	}
	return BitIterator{nIter.nibblePath, nIter.start*4, nIter.end*4}
}

// TODO: UNDERSTAND CHAIN
// get all nibblePath
func (nIter *NibbleIterator)getNibblePath() NibblePath {
	return nIter.nibblePath
}

// get nibblePath based end
func (nIter NibbleIterator)getPartNibblePath() NibblePath {
	partNibblePath := NibblePath{}
	for i:=nIter.start; i<nIter.end; i++ {
		partNibblePath.push(nIter.nibblePath.getNibble(i))
	}
	return partNibblePath
}

// Get the number of nibbles that this iterator covers.
func (nIter NibbleIterator)numNibbles() uint {
	if nIter.start > nIter.end {
		panic("out of range")
	}
	return uint(nIter.end-nIter.start)
}

func (nIter *NibbleIterator)isFinished() bool {
	return nIter.peek() == nil
}

/// Advance both iterators if their next nibbles are the same until either reaches the end or
/// the find a mismatch. Return the number of matched nibbles.
func SkipCommonPrefix(x interator, y interator) uint {
	var count uint = 0
	for {
		xPeek := x.peek()
		yPeek := y.peek()
		if xPeek == nil || yPeek == nil || xPeek != yPeek {
			break
		}
		count += 1
		x.next()
		y.next()
	}
	return count
}

