package core

import (
	"crypto/md5"
	"github.com/pkg/errors"
	"github.com/ss_go/shadowaead"
	"net"
	"strings"
)

type Cipher interface {
	StreamConnCipher
}

type aeadCipher struct {
	shadowaead.Cipher
}

type StreamConnCipher interface {
	StreamConn(net.Conn) net.Conn
}

func (aead *aeadCipher) StreamConn(c net.Conn) net.Conn {
	return shadowaead.NewConn(c, aead)
}

var aeadList = map[string]struct {
	Keysize int
	New     func([]byte) (shadowaead.Cipher, error)
}{
	"AEAD_CHACHA20_POLY1305": {32, shadowaead.Chacha20Poly1305},
}

func PickCipher(name string, password string) (Cipher, error) {
	name = strings.ToUpper(name)
	if choice, ok := aeadList[name]; ok {
		key := kdf(password, choice.Keysize)
		adad, err := choice.New(key)
		return &aeadCipher{adad}, err
	}
	return nil, errors.New("chipher not supported")
}

func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}
