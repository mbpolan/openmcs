package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BitSet_Empty(t *testing.T) {
	b := NewBitSet()

	assert.Empty(t, b.bytes)
}

func Test_BitSet_Set(t *testing.T) {
	b := NewBitSet()

	b.Set()

	assert.Equal(t, 1, len(b.bytes))
	assert.Equal(t, uint8(0x80), b.bytes[0])
}

func Test_BitSet_SetBits(t *testing.T) {
	b := NewBitSet()

	b.SetBits(0x7FF, 11)

	assert.Equal(t, 2, len(b.bytes))
	assert.Equal(t, uint8(0xFF), b.bytes[0])
	assert.Equal(t, uint8(0xE0), b.bytes[1])
}

func Test_BitSet_Skip(t *testing.T) {
	b := NewBitSet()

	b.Skip()

	assert.Equal(t, 1, len(b.bytes))
	assert.Equal(t, uint8(0x00), b.bytes[0])
}
