package operator

import (
	"crypto/ecdsa"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// PrivateKey represents peer's static key associated with an on-chain
// stake. It is used to authenticate the peer and for attributability (signing).
type PrivateKey = ecdsa.PrivateKey

// PublicKey represents peer's static key associated with an on-chain
// stake. It is used to authenticate the peer and for attributability
// (verification).
type PublicKey = ecdsa.PublicKey

// GenerateKeyPair generates a new, random static key based on
// secp256k1 ethereum curve.
func GenerateKeyPair(rand io.Reader) (*PrivateKey, *PublicKey, error) {
	ecdsaKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand)
	if err != nil {
		return nil, nil, err
	}

	return (*PrivateKey)(ecdsaKey), (*PublicKey)(&ecdsaKey.PublicKey), nil
}

// EthereumKeyToOperatorKey transforms a `go-ethereum`-based ECDSA key into the
// format supported by all packages used in keep-core.
func EthereumKeyToOperatorKey(ethereumKey *keystore.Key) (*PrivateKey, *PublicKey) {
	privKey := ethereumKey.PrivateKey
	return (*PrivateKey)(privKey), (*PublicKey)(&privKey.PublicKey)
}

// Sign calculates an ECDSA signature. Modified code from go-ethereum.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	sig, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(prv), hash, false)
	if err != nil {
		return nil, err
	}
	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[64] = v
	return sig, nil
}

// VerifySignature checks that the given pubkey created signature over message.
// The public key should be in compressed (33 bytes) or uncompressed (65 bytes) format.
// The signature should be in [R || S] format.
func VerifySignature(publicKey *PublicKey, hash, signature []byte) error {
	// Convert the operator's static key into an uncompressed public key
	// which should be 65 bytes in length.
	uncompressedPubKey := (*btcec.PublicKey)(publicKey).SerializeUncompressed()

	// If our signature is in the [R || S || V] format, ensure we strip out
	// the Ethereum-specific recovery-id, V, if it already hasn't been done.
	if len(signature) == 65 {
		signature = signature[:len(signature)-1]
	}

	// The signature should be 64 bytes.
	if len(signature) != 64 {
		return fmt.Errorf(
			"malformed signature %+v with length %d",
			signature,
			len(signature),
		)
	}

	if verified := crypto.VerifySignature(
		uncompressedPubKey,
		hash,
		signature,
	); !verified {
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}
