package types

import (
	"encoding/binary"
	"fmt"
	"hash"
	"reflect"

	"github.com/daragao/merkletree"
	"github.com/ethereum/go-ethereum/crypto"
)

// Token fully defines an collection token by specifying its tokenID and features.
type Token struct {
	TokenID  uint16
	Features []uint8
}

// Encode encodes the token as field by serialising the features
func (f Token) Encode() ([]byte, error) {
	return []byte(f.Features), nil
}

// Label labels each token with its tokenID.
// Needed for the use with labelled buckets.
func (f Token) Label() uint16 {
	return f.TokenID
}

// CalculateHash calculates the hash of a token.
// Needed in the computation of merkle trees.
func (f Token) CalculateHash() ([]byte, error) {
	tmp := make([]byte, 64)

	// | 0..0 (30 bytes) | tokenId (2 bytes) | features (32 bytes) |
	binary.BigEndian.PutUint16(tmp[30:], f.TokenID)
	copy(tmp[64-len(f.Features):], f.Features)

	return crypto.Keccak256(tmp), nil
}

// Equals checks if two tokens are equal.
// Needed in the computation of merkle trees.
func (f Token) Equals(other merkletree.Content) (bool, error) {
	o, ok := other.(Token)
	if !ok {
		return false, fmt.Errorf("cannot convert %T to %T", other, f)
	}
	return reflect.DeepEqual(f, o), nil
}

// ComputeMerkleTree computes the merkle tree for a list of tokens. This routine
// uses keccak256 for hashing and sorts leafs by size before doing so to be
// compatible with OpenZeppelin's merkle proof validator in Solidity.
func ComputeMerkleTree(tokens []Token) (*merkletree.MerkleTree, error) {
	c := make([]merkletree.Content, len(tokens))
	for i, v := range tokens {
		c[i] = v
	}

	mt, err := merkletree.NewTreeWithHashStrategySorted(c, func() hash.Hash {
		return crypto.NewKeccakState()
	}, true)
	if err != nil {
		return nil, fmt.Errorf("merkletree.NewTreeWithHashStrategySorted([data], crypto.NewKeccakState, true): %v", err)
	}

	return mt, nil
}
