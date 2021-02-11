package jellyfish

import "fmt"

type NibblePath struct {
	NumNibbles int
	Bytes      []uint8
}

type Nibble = uint8

// Creates a new `NibblePath` from a vector of Bytes assuming each byte has 2 nibbles.
func (np *NibblePath) new(bytes []uint8) (NibblePath, error) {
	// TODO: Bytes len check
	numNibbles := len(bytes) * 2
	return NibblePath{numNibbles, bytes}, nil
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

func (np NibblePath) bits() {

}

func (np *NibblePath) nibbles() {

}
