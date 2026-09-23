package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBcryptHasherRoundTripAndMismatch(t *testing.T) {
	hasher := NewBcryptHasher()

	hash, err := hasher.Hash("correct horse battery staple")

	require.NoError(t, err)
	require.NotEqual(t, "correct horse battery staple", hash)
	require.True(t, hasher.Compare(hash, "correct horse battery staple"))
	require.False(t, hasher.Compare(hash, "wrong password"))
	require.False(t, hasher.Compare("not-a-bcrypt-hash", "correct horse battery staple"))
}

func TestBcryptHasherRejectsPasswordsOverSeventyTwoBytes(t *testing.T) {
	_, err := NewBcryptHasher().Hash(strings.Repeat("a", 73))

	require.Error(t, err)
}
