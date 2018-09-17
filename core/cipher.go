package core

import (
	"crypto/md5"
	"errors"
	"github.com/ss_go/shadowaead"
	"net"
	"strings"
)

type Cipher interface {
	StreamConnCipher
}

// ErrCipherNotSupported occurs when a cipher is not supported (likely because of security concerns).
var ErrCipherNotSupported = errors.New("cipher not supported")

type aeadCipher struct{
	shadowaead.Cipher
}
func (aead *aeadCipher) StreamConn(c net.Conn) net.Conn { return shadowaead.NewConn(c, aead) }
type StreamConnCipher interface {
	StreamConn(net.Conn) net.Conn
}
type dummy struct{}
func (dummy) StreamConn(c net.Conn) net.Conn { return c }

var aeadList=map[string]struct{
	Keysize int
	New func([]byte)(shadowaead.Cipher,error)
}{
	"AEAD_AES_128_GCM":       {16, shadowaead.AESGCM},
	"AEAD_AES_192_GCM":       {24, shadowaead.AESGCM},
	"AEAD_AES_256_GCM":       {32, shadowaead.AESGCM},
	"AEAD_CHACHA20_POLY1305": {32, shadowaead.Chacha20Poly1305},

}



func PickCipher(name string,key []byte,password string)(Cipher,error)  {
	name = strings.ToUpper(name)

	switch name {
	case "DUMMY":
		return &dummy{},nil
	case "CHACHA20-IETF-POLY1305":
		name = "AEAD_CHACHA20_POLY1305"
	case "AES-128-GCM":
		name = "AEAD_AES_128_GCM"
	case "AES-196-GCM":
		name = "AEAD_AES_196_GCM"
	case "AES-256-GCM":
		name = "AEAD_AES_256_GCM"
	}

	if choice,ok :=aeadList[name];ok{
		if len(key) ==0{
			key = kdf(password,choice.Keysize)
		}
		aead, err := choice.New(key)
		return &aeadCipher{aead}, err

	}
	return nil, ErrCipherNotSupported

}
// key-derivation function from original Shadowsocks
func kdf(password string,keyLen int) []byte  {
	var b,prev []byte
	h:=md5.New()
	for len(b)<keyLen{
		h.Write(prev)
		h.Write([]byte(password))
		b=h.Sum(b)
		prev=b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}