package cert

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
)

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int
	Exp   *big.Int
	Coeff *big.Int
}

type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	Dp      *big.Int `asn1:"optional"`
	Dq      *big.Int `asn1:"optional"`
	Qinv    *big.Int `asn1:"optional"`
	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}
