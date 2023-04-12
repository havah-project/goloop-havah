package hvhstate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToKey(t *testing.T) {
	addr := newDummyAddress(1, false)
	key := ToKey(addr)
	assert.Equal(t, string(addr.Bytes()), key)

	key = ToKey("hello")
	assert.Equal(t, "hello", key)

	o := []byte{0, 1, 2, 3}
	key = ToKey(o)
	assert.Equal(t, string(o), key)
}

func TestIsItTimeToCheckBlockVote(t *testing.T) {
	tests := []struct{
		blockIndex int64
		period int64
		expResult bool
	}{
		{0, 0, false},
		{50, 0, false},
		{0, 100, true},
		{1, 100, false},
		{100, 100, true},
		{150, 100, false},
		{200, 100, true},
	}
	for i, test := range tests {
		name := fmt.Sprintf("name-%d", i)
		t.Run(name, func(t *testing.T) {
			ok := IsItTimeToCheckBlockVote(test.blockIndex, test.period)
			assert.Equal(t, test.expResult, ok)
		})
	}
}

func TestGetTermStartAndIndex(t *testing.T) {
	tests := []struct{
		// input
		height int64
		issueStart int64
		termPeriod int64
		// output
		termStart int64
		termIndex int64
		success bool
	}{
		{10, 20, 30, -1, -1, false},
		{17, 17, 100, 17, 0, true},
		{11, 10, 100, 10, 1, true},
		{110, 10, 100, 110, 0, true},
		{111, 10, 100, 110, 1, true},
	}

	for i, test := range tests {
		name := fmt.Sprintf("name-%02d", i)
		t.Run(name, func(t *testing.T){
			termStart, termIndex, err := GetTermStartAndIndex(
				test.height, test.issueStart, test.termPeriod)
			assert.Equal(t, test.termStart, termStart)
			assert.Equal(t, test.termIndex, termIndex)
			assert.Equal(t, test.success, err == nil)
		})
	}
}
