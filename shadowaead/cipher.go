package shadowaead

import (
	"crypto/cipher"
	"crypto/sha1"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
	"io"
	"strconv"
)

type KeySizeError int
func (e KeySizeError) Error() string {
	return "key size error: need " + strconv.Itoa(int(e)) + " bytes"
}


type Cipher interface {
	KeySize() int
	SaltSize() int
	Encrypter(salt []byte) (cipher.AEAD, error)
	Decrypter(salt []byte) (cipher.AEAD, error)
}

func (a *metaCipher) KeySize() int { return len(a.psk) }
func (a *metaCipher) SaltSize() int {
	if ks := a.KeySize(); ks > 16 {
		return ks
	}
	return 16
}
func (a *metaCipher) Encrypter(salt []byte) (cipher.AEAD, error) {
	subkey := make([]byte, a.KeySize())
	hkdfSHA1(a.psk, salt, []byte("ss-subkey"), subkey)
	return a.makeAEAD(subkey)
}
func (a *metaCipher) Decrypter(salt []byte) (cipher.AEAD, error) {
	subkey := make([]byte, a.KeySize())
	hkdfSHA1(a.psk, salt, []byte("ss-subkey"), subkey)
	return a.makeAEAD(subkey)
}


func hkdfSHA1(secret, salt, info, outkey []byte) {
	r := hkdf.New(sha1.New, secret, salt, info)
	if _, err := io.ReadFull(r, outkey); err != nil {
		panic(err) // should never happen
	}
}
type metaCipher struct {
	psk      []byte
	makeAEAD func(key []byte) (cipher.AEAD, error)
}
// Chacha20Poly1305 creates a new Cipher with a pre-shared key. len(psk)
// must be 32.
func Chacha20Poly1305(psk []byte) (Cipher, error) {
	if len(psk) != chacha20poly1305.KeySize {
		return nil, KeySizeError(chacha20poly1305.KeySize)
	}
	return &metaCipher{psk: psk, makeAEAD: chacha20poly1305.New}, nil
}
