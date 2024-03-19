package consensus

import (
	"fmt"
	"math/bits"
	"math/rand"

	"github.com/icon-project/goloop/common/errors"
)

type word = uint64

const wordBits = 64

type BitArray struct {
	NumBits int
	Words   []word
}

func (ba *BitArray) Verify() error {
	if ba.NumBits > len(ba.Words)*wordBits {
		return errors.Errorf("invalid BitArray NumBits=%d Words.len=%d", ba.NumBits, len(ba.Words))
	}
	return nil
}

func (ba *BitArray) Len() int {
	return ba.NumBits
}

func (ba *BitArray) Set(idx int) {
	if idx >= ba.NumBits {
		return
	}
	ba.Words[idx/wordBits] = ba.Words[idx/wordBits] | (1 << uint(idx%wordBits))
}

func (ba *BitArray) Unset(idx int) {
	if idx >= ba.NumBits {
		return
	}
	ba.Words[idx/wordBits] = ba.Words[idx/wordBits] &^ (1 << uint(idx%wordBits))
}

func (ba *BitArray) Put(idx int, v bool) {
	if idx >= ba.NumBits {
		return
	}
	if v {
		ba.Set(idx)
	} else {
		ba.Unset(idx)
	}
}

func (ba *BitArray) Get(idx int) bool {
	if idx >= ba.NumBits {
		return false
	}
	return ba.Words[idx/wordBits]&(1<<uint(idx%wordBits)) != 0
}

func (ba *BitArray) Flip() {
	l := len(ba.Words)
	for i := 0; i < l; i++ {
		ba.Words[i] = ^ba.Words[i]
	}
	if l > 0 {
		ba.Words[l-1] = ba.Words[l-1] & ((1 << uint(ba.NumBits%wordBits)) - 1)
	}
}

func (ba *BitArray) AssignAnd(ba2 *BitArray) {
	lba := len(ba.Words)
	lba2 := len(ba2.Words)
	if ba.NumBits > ba2.NumBits {
		ba.Words = ba.Words[:lba2]
		ba.NumBits = ba2.NumBits
		lba = lba2
	}
	for i := 0; i < lba; i++ {
		ba.Words[i] &= ba2.Words[i]
	}
}

func (ba *BitArray) PickRandom() int {
	var count int
	for i := 0; i < len(ba.Words); i++ {
		count = count + bits.OnesCount64(ba.Words[i])
	}
	if count == 0 {
		return -1
	}
	pick := rand.Intn(count)
	for i := 0; i < len(ba.Words); i++ {
		c := bits.OnesCount64(ba.Words[i])
		if pick < c {
			for idx := i * wordBits; idx < ba.NumBits; idx++ {
				if ba.Get(idx) {
					if pick == 0 {
						return idx
					}
					pick--
				}
			}
		}
		pick = pick - c
	}
	panic("PickRandom: internal error")
}

func (ba BitArray) String() string {
	// TODO better form?
	return fmt.Sprintf("%x", ba.Words)
}

func (ba *BitArray) Equal(ba2 *BitArray) bool {
	lba := len(ba.Words)
	if ba.NumBits != ba2.NumBits {
		return false
	}
	for i := 0; i < lba; i++ {
		if ba.Words[i] != ba2.Words[i] {
			return false
		}
	}
	return true
}

func (ba *BitArray) Copy() *BitArray {
	ba2 := NewBitArray(ba.NumBits)
	copy(ba2.Words, ba.Words)
	return ba2
}

func NewBitArray(n int) *BitArray {
	return &BitArray{n, make([]word, (n+wordBits-1)/wordBits)}
}
