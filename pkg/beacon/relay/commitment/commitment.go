// Package commitment implements a commitment scheme described by
// Torben Pryds Pedersen in the referenced [Ped] paper.
//
// [Ped] Pedersen T.P. (1992) Non-Interactive and Information-Theoretic Secure
// Verifiable Secret Sharing. In: Feigenbaum J. (eds) Advances in Cryptology —
// CRYPTO ’91. CRYPTO 1991. Lecture Notes in Computer Science, vol 576. Springer,
// Berlin, Heidelberg
package commitment

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/keep-network/keep-core/pkg/internal/byteutils"
)

// Parameters specific to the scheme
type Parameters struct {
	p *big.Int // Prime such that p = 2q + 1
	q *big.Int // Sophie Germain prime

	// Elements of a subgroup of quadratic residues of order q
	// g,h are elements of a group of order q such that nobody knows log_g(h)
	g, h *big.Int
}

// Commitment is produced for each message we have committed to.
// It is usually revealed to the verifier immediately after it has been produced
// and lets to verify if the message revealed later by the committing party
// is really what that party has committed to.
//
// The commitment itself is not enough for a verification. In order to perform
// a verification, the interested party must receive the `DecommitmentKey`.
type Commitment struct {
	commitment *big.Int
}

// DecommitmentKey allows to open a commitment and verify if the value is what
// we have really committed to.
type DecommitmentKey struct {
	r *big.Int
}

// GenerateParameters generates parameters for a scheme execution
//
// TODO g and h shouldn't be generated by the committer. We need to decide how to
// do this, so nobody knows log_g(h)
func GenerateParameters() (*Parameters, error) {
	parameters, err := initializeParameters()
	if err != nil {
		return nil, fmt.Errorf("parameters initialization failed [%s]", err)
	}

	randomG, err := randomFromZn(parameters.p)
	if err != nil {
		return nil, fmt.Errorf("g generation failed [%s]", err)
	}
	parameters.g = new(big.Int).Exp(randomG, big.NewInt(2), nil) // (randomZ(0, 2^p - 1]) ^2

	randomH, err := randomFromZn(parameters.p)
	if err != nil {
		return nil, fmt.Errorf("h generation failed [%s]", err)
	}
	// TODO h can be jointly calculated by players with distributed coin flipping protocol
	parameters.h = new(big.Int).Exp(randomH, big.NewInt(2), nil) // (randomZ(0, 2^p - 1]) ^2

	return parameters, nil
}

// Generate evaluates a commitment and a decommitment key with specific master
// public key for the secret messages provided as an argument.
func Generate(parameters *Parameters, secret []byte) (*Commitment, *DecommitmentKey, error) {
	r, err := randomFromZn(parameters.q) // randomZ(0, 2^q - 1]
	if err != nil {
		return nil, nil, fmt.Errorf("r generation failed [%s]", err)
	}

	digest := calculateDigest(secret, parameters.q)

	commitment := calculateCommitment(parameters, digest, r)

	return &Commitment{commitment},
		&DecommitmentKey{r},
		nil
}

// Verify checks the received commitment against the revealed secret message.
func (c *Commitment) Verify(parameters *Parameters, decommitmentKey *DecommitmentKey, secret []byte) bool {
	digest := calculateDigest(secret, parameters.q)
	expectedCommitment := calculateCommitment(parameters, digest, decommitmentKey.r)
	return expectedCommitment.Cmp(c.commitment) == 0
}

func calculateDigest(secret []byte, mod *big.Int) *big.Int {
	hash := byteutils.Sha256Sum(secret)
	digest := new(big.Int).Mod(hash, mod)
	return digest
}

func calculateCommitment(parameters *Parameters, digest, r *big.Int) *big.Int {
	// ((g ^ digest) % p) * ((h ^ r) % p)
	return new(big.Int).Mul(
		new(big.Int).Exp(parameters.g, digest, parameters.p),
		new(big.Int).Exp(parameters.h, r, parameters.p),
	)
}

// initializeParameters sets p and q to predefined fixed values,
// such that `p = 2q + 1`.
// - `p` is 4096-bit safe prime
// - `q` is 4095-bit Sophie Germain prime
func initializeParameters() (*Parameters, error) {
	pStr := "0xc8526644a9c4739683742b7003640b2023ca42cc018a42b02a551bb825c6828f86e2e216ea5d31004c433582a3fa720459efb42e091d73fb281810e1825691f0799811be62ae57f62ab00670edd35426d108d3b9c4fd008eddc67275a0489fe132e4c31bd7069ea7884cbb8f8f9255fe7b87fc0099f246776c340912df48f7945bc2bc0bc6814978d27b7af2ebc41f458ae795186db0fd7e6151bb8a7fe2b41370f7a2848ef75d3ec88f3439022c10e78b434c2f24b2f40bd02930e6c8aadef87b0dc87cdba07dcfa86884a168bd1381a4f48be12e5d98e41f954c37aec011cc683570e8890418756ed98ace8c8e59ae1df50962c1622fe66b5409f330cad6b7c68f2e884786d9807190b89ac4a3b3507e49b2dd3f33d765ad29e2015180c8cd0258dd8bdaab17be5d74871fec04c492240c6a2692b2c9a62c9adbaac34a333f135801ff948e8dfb6bbd6212a67950fb8edd628d05d19d1b94e9be7c52ed484831d50adaa29e71de197e351878f1c40ec67ee809e824124529e27bd5ecf3054f6784153f7db27ff0c87420bb2b2754ed363fc2ba8399d49d291f342173e7619183467a9694efa243e1d41b26c13b38ca0f43bb7c9050eb966461f28436583a9d13d2c1465b78184eae360f009505ccea288a053d111988d55c12befd882a857a530efac2c0592987cd83c39844a10e058739ab1c39006a3123e7fc887845675f"
	p, result := new(big.Int).SetString(pStr, 0)
	if !result {
		return nil, fmt.Errorf("converting p failed")
	}

	qStr := "0x6429332254e239cb41ba15b801b2059011e5216600c52158152a8ddc12e34147c371710b752e988026219ac151fd39022cf7da17048eb9fd940c0870c12b48f83ccc08df31572bfb1558033876e9aa13688469dce27e80476ee3393ad0244ff09972618deb834f53c4265dc7c7c92aff3dc3fe004cf9233bb61a04896fa47bca2de15e05e340a4bc693dbd7975e20fa2c573ca8c36d87ebf30a8ddc53ff15a09b87bd142477bae9f64479a1c81160873c5a1a61792597a05e814987364556f7c3d86e43e6dd03ee7d4344250b45e89c0d27a45f0972ecc720fcaa61bd76008e6341ab87444820c3ab76cc56746472cd70efa84b160b117f335aa04f998656b5be347974423c36cc038c85c4d6251d9a83f24d96e9f99ebb2d694f100a8c06466812c6ec5ed558bdf2eba438ff602624912063513495964d3164d6dd561a5199f89ac00ffca4746fdb5deb109533ca87dc76eb14682e8ce8dca74df3e2976a42418ea856d514f38ef0cbf1a8c3c78e207633f7404f412092294f13deaf67982a7b3c20a9fbed93ff8643a105d9593aa769b1fe15d41ccea4e948f9a10b9f3b0c8c1a33d4b4a77d121f0ea0d93609d9c6507a1ddbe482875cb3230f9421b2c1d4e89e960a32dbc0c27571b07804a82e6751445029e888cc46aae095f7ec41542bd29877d61602c94c3e6c1e1cc22508702c39cd58e1c80351891f3fe443c22b3af"
	q, result := new(big.Int).SetString(qStr, 0)
	if !result {
		return nil, fmt.Errorf("converting q failed")
	}

	return &Parameters{p: p, q: q}, nil
}

// randomFromZn generates a random `big.Int` in a range (0, 2^n - 1]
func randomFromZn(n *big.Int) (*big.Int, error) {
	x := big.NewInt(0)
	var err error
	// 2^n - 1
	// TODO check if this is what we really need for g,h and r
	max := new(big.Int).Sub(
		new(big.Int).Exp(big.NewInt(2), n, nil),
		big.NewInt(1),
	)
	for x.Sign() == 0 {
		x, err = rand.Int(rand.Reader, max)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random number [%s]", err)
		}
	}
	return x, nil
}
