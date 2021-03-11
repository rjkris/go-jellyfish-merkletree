package jellyfish

import (
	"fmt"
	"testing"
)

func TestPush(t *testing.T)  {
	path1 := NibblePath{}.new([]uint8{1})
	fmt.Printf("before: %+v", path1)
	path1.push(Nibble(2))
	fmt.Printf("after: %+v", path1)
	path1.push(Nibble(3))
	fmt.Printf("after: %+v", path1)
}
