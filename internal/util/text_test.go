package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DecodeName_valid(t *testing.T) {
	encoded := EncodeName("abcdefghijkl")

	decoded, err := DecodeName(encoded)

	assert.NoError(t, err)
	assert.Equal(t, "abcdefghijkl", decoded)
}
