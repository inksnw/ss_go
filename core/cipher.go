package core

import (
	"crypto/md5"
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

var aeadList = map[string]struct {
	KeySize int
	New     func([]byte) (shadowaead.Cipher, error)
}{
	"AEAD_CHACHA20_POLY1305": {32, shadowaead.Chacha20Poly1305},
}

func (aead *aeadCipher) StreamConn(c net.Conn) net.Conn {
	return shadowaead.NewConn(c, aead)
}

func PickCipher(name, password string) Cipher {
	name = strings.ToUpper(name)
	if choice, ok := aeadList[name]; ok {
		key := kdf(password, choice.KeySize)
		aead, _ := choice.New(key)
		return &aeadCipher{aead}
	}
	return nil

}

func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
	}
	return b[:keyLen]
}
