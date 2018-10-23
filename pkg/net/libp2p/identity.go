package libp2p

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/keep-network/keep-core/pkg/chain/ethereum"
	"github.com/keep-network/keep-core/pkg/net/gen/pb"

	libp2pcrypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// identity represents a group member's network level identity. It
// implements the net.TransportIdentifier interface. A valid group member will
// generate or provide a keypair, which will correspond to a network ID.
//
// Consumers of the net package require an ID to register with protocol level
// IDs, as well as a public key for authentication.
type identity struct {
	id      peer.ID
	pubKey  libp2pcrypto.PubKey
	privKey libp2pcrypto.PrivKey
}

type networkIdentity peer.ID

func (networkIdentity) ProviderName() string {
	return "libp2p"
}

func (ni networkIdentity) String() string {
	return peer.ID(ni).Pretty()
}

func (i *identity) Marshal() ([]byte, error) {
	var (
		err error
	)

	pubKey := i.pubKey
	if pubKey == nil {
		pubKey, err = i.id.ExtractPublicKey()
		if err != nil {
			return nil, err
		}
	}
	if pubKey == nil {
		return nil, fmt.Errorf(
			"failed to generate public key with peerid %v",
			i.id.Pretty(),
		)
	}
	pubKeyBytes, err := pubKey.Bytes()
	if err != nil {
		return nil, err
	}
	return (&pb.Identity{PubKey: pubKeyBytes}).Marshal()
}

func (i *identity) Unmarshal(bytes []byte) error {
	var (
		err        error
		pid        peer.ID
		pbIdentity pb.Identity
	)

	if err = pbIdentity.Unmarshal(bytes); err != nil {
		return fmt.Errorf("unmarshalling failed with error %s", err)
	}
	i.pubKey, err = libp2pcrypto.UnmarshalPublicKey(pbIdentity.PubKey)
	if err != nil {
		return err
	}
	pid, err = peer.IDFromPublicKey(i.pubKey)
	if err != nil {
		return fmt.Errorf("Failed to generate valid libp2p identity with err: %s", err)
	}
	i.id = pid

	return nil
}

func readSecp256k1Key(account ethereum.Account) (
	libp2pcrypto.PrivKey,
	libp2pcrypto.PubKey,
	error,
) {
	ethereumKey, err := ethereum.DecryptKeyFile(
		account.KeyFile,
		account.KeyFilePassword,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read KeyFile: %s [%v]", account.KeyFile, err,
		)
	}

	privKey, _ := btcec.PrivKeyFromBytes(
		btcec.S256(), ethereumKey.PrivateKey.D.Bytes(),
	)

	k := (*libp2pcrypto.Secp256k1PrivateKey)(privKey)
	return k, k.GetPublic(), nil
}

func readIdentity(account ethereum.Account) (*identity, error) {
	privateKey, publicKey, err := readSecp256k1Key(account)
	if err != nil {
		return nil, fmt.Errorf("could not load peer's static key [%v]", err)
	}

	peerID, err := peer.IDFromPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf(
			"could not transform public key to peer's identity [%v]", err,
		)
	}

	return &identity{peerID, publicKey, privateKey}, nil
}
