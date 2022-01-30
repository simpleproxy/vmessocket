package crypto

import (
	"crypto/cipher"

	"github.com/vmessocket/vmessocket/common/crypto/internal"
)

func NewChaCha20Stream(key []byte, iv []byte) cipher.Stream {
	return internal.NewChaCha20Stream(key, iv, 20)
}
