package chain

import (
	"bytes"
	"fmt"
)

// DKGResult is a result of distributed key generation protocol.
//
// If the protocol execution finishes with an acceptable number of disqualified
// or inactive members, the group with remaining list of honest members will
// be added to the signing groups list for the threshold relay.
//
// Otherwise, group creation will not finish, which will be due to either the
// number of inactive or disqualified participants, or the results (signatures)
// being disputed in a way where the correct outcome cannot be ascertained.
type DKGResult struct {
	// Group public key generated by the protocol execution, empty if the protocol failed.
	GroupPublicKey []byte
	// Misbehaved members are all members either inactive or disqualified.
	// Misbehaved members are represented as a slice of bytes for optimizing
	// on-chain storage. Each byte is an inactive or disqualified member index.
	Misbehaved []byte
}

// DKGResultHash is a 256-bit hash of DKG Result. The hashing algorithm should
// be the same as the one used on-chain.
type DKGResultHash [hashByteSize]byte

const hashByteSize = 32

// DKGResultsVotes is a map of votes for each DKG Result.
type DKGResultsVotes map[DKGResultHash]int

// Equals checks if two DKG results are equal.
func (r *DKGResult) Equals(r2 *DKGResult) bool {
	if r == nil || r2 == nil {
		return r == r2
	}
	if !bytes.Equal(r.GroupPublicKey, r2.GroupPublicKey) {
		return false
	}
	if !bytes.Equal(r.Misbehaved, r2.Misbehaved) {
		return false
	}

	return true
}

// DKGResultHashFromBytes converts bytes slice to DKG Result Hash. It requires
// provided bytes slice size to be exactly 32 bytes.
func DKGResultHashFromBytes(bytes []byte) (DKGResultHash, error) {
	var hash DKGResultHash

	if len(bytes) != hashByteSize {
		return hash, fmt.Errorf("bytes length is not equal %v", hashByteSize)
	}
	copy(hash[:], bytes[:])

	return hash, nil
}
