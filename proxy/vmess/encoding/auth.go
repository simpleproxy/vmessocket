package encoding

import (
	"crypto/md5"
	"encoding/binary"
	"hash/fnv"

	"golang.org/x/crypto/sha3"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/crypto"
)

type AEADSizeParser struct {
	crypto.AEADChunkSizeParser
}

type FnvAuthenticator struct{}

type NoOpAuthenticator struct{}

type ShakeSizeParser struct {
	shake  sha3.ShakeHash
	buffer [2]byte
}

func Authenticate(b []byte) uint32 {
	fnv1hash := fnv.New32a()
	common.Must2(fnv1hash.Write(b))
	return fnv1hash.Sum32()
}

func GenerateChacha20Poly1305Key(b []byte) []byte {
	key := make([]byte, 32)
	t := md5.Sum(b)
	copy(key, t[:])
	t = md5.Sum(key[:16])
	copy(key[16:], t[:])
	return key
}

func NewAEADSizeParser(auth *crypto.AEADAuthenticator) *AEADSizeParser {
	return &AEADSizeParser{crypto.AEADChunkSizeParser{Auth: auth}}
}

func NewShakeSizeParser(nonce []byte) *ShakeSizeParser {
	shake := sha3.NewShake128()
	common.Must2(shake.Write(nonce))
	return &ShakeSizeParser{
		shake: shake,
	}
}

func (s *ShakeSizeParser) Decode(b []byte) (uint16, error) {
	mask := s.next()
	size := binary.BigEndian.Uint16(b)
	return mask ^ size, nil
}

func (s *ShakeSizeParser) Encode(size uint16, b []byte) []byte {
	mask := s.next()
	binary.BigEndian.PutUint16(b, mask^size)
	return b[:2]
}

func (s *ShakeSizeParser) MaxPaddingLen() uint16 {
	return 64
}

func (s *ShakeSizeParser) next() uint16 {
	common.Must2(s.shake.Read(s.buffer[:]))
	return binary.BigEndian.Uint16(s.buffer[:])
}

func (s *ShakeSizeParser) NextPaddingLen() uint16 {
	return s.next() % 64
}

func (*FnvAuthenticator) NonceSize() int {
	return 0
}

func (NoOpAuthenticator) NonceSize() int {
	return 0
}

func (*FnvAuthenticator) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if binary.BigEndian.Uint32(ciphertext[:4]) != Authenticate(ciphertext[4:]) {
		return dst, newError("invalid authentication")
	}
	return append(dst, ciphertext[4:]...), nil
}

func (NoOpAuthenticator) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	return append(dst[:0], ciphertext...), nil
}

func (*FnvAuthenticator) Overhead() int {
	return 4
}

func (NoOpAuthenticator) Overhead() int {
	return 0
}

func (*FnvAuthenticator) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	dst = append(dst, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(dst, Authenticate(plaintext))
	return append(dst, plaintext...)
}

func (NoOpAuthenticator) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	return append(dst[:0], plaintext...)
}

func (*ShakeSizeParser) SizeBytes() int32 {
	return 2
}
