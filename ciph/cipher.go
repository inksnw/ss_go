package ciph

import (
	"crypto/md5"
	"github.com/shadowsocks/go-shadowsocks2/shadowstream"
	"github.com/ss_go/shadowaead"
	"net"
	"strings"
)

type Cipher interface {
	StreamConnCipher
	PacketConnCipher
}

type StreamConnCipher interface {
	StreamConn(net.Conn) net.Conn
}

type PacketConnCipher interface {
	PacketConn(net.PacketConn) net.PacketConn
}

const (
	aeadAes128Gcm        = "AEAD_AES_128_GCM"
	aeadAes192Gcm        = "AEAD_AES_192_GCM"
	aeadAes256Gcm        = "AEAD_AES_256_GCM"
	aeadChacha20Poly1305 = "AEAD_CHACHA20_POLY1305"
)

var aeadList = map[string]struct {
	KeySize int
	New     func([]byte) (shadowaead.Cipher, error)
}{
	aeadAes128Gcm:        {16, shadowaead.AESGCM},
	aeadAes192Gcm:        {24, shadowaead.AESGCM},
	aeadAes256Gcm:        {32, shadowaead.AESGCM},
	aeadChacha20Poly1305: {32, shadowaead.Chacha20Poly1305},
}

// List of stream ciphers: key size in bytes and constructor
var streamList = map[string]struct {
	KeySize int
	New     func(key []byte) (shadowstream.Cipher, error)
}{
	"AES-128-CTR":   {16, shadowstream.AESCTR},
	"AES-192-CTR":   {24, shadowstream.AESCTR},
	"AES-256-CTR":   {32, shadowstream.AESCTR},
	"AES-128-CFB":   {16, shadowstream.AESCFB},
	"AES-192-CFB":   {24, shadowstream.AESCFB},
	"AES-256-CFB":   {32, shadowstream.AESCFB},
	"CHACHA20-IETF": {32, shadowstream.Chacha20IETF},
	"XCHACHA20":     {32, shadowstream.Xchacha20},
}

func PickCipher(name string, password string) (Cipher, error) {
	name = strings.ToUpper(name)
	switch name {
	case "DUMMY":
		return &dummy{}, nil
	case "CHACHA20-IETF-POLY1305":
		name = aeadChacha20Poly1305
	case "AES-128-GCM":
		name = aeadAes128Gcm
	case "AES-192-GCM":
		name = aeadAes192Gcm
	case "AES-256-GCM":
		name = aeadAes256Gcm
	}
	if choice, ok := aeadList[name]; ok {
		key := kdf(password, choice.KeySize)
		aead, err := choice.New(key)
		return &aeadCipher{aead}, err

	}

	if choice, ok := streamList[name]; ok {

		key := kdf(password, choice.KeySize)

		ciph, err := choice.New(key)
		return &streamCipher{ciph}, err
	}

	return &dummy{}, nil
}

type aeadCipher struct{ shadowaead.Cipher }

func (aead *aeadCipher) StreamConn(c net.Conn) net.Conn { return shadowaead.NewConn(c, aead) }
func (aead *aeadCipher) PacketConn(c net.PacketConn) net.PacketConn {
	return shadowaead.NewPacketConn(c, aead)
}

type dummy struct{}

func (dummy) StreamConn(c net.Conn) net.Conn             { return c }
func (dummy) PacketConn(c net.PacketConn) net.PacketConn { return c }

type streamCipher struct{ shadowstream.Cipher }

func (ciph *streamCipher) StreamConn(c net.Conn) net.Conn { return shadowstream.NewConn(c, ciph) }
func (ciph *streamCipher) PacketConn(c net.PacketConn) net.PacketConn {
	return shadowstream.NewPacketConn(c, ciph)
}

// key-derivation function from original Shadowsocks
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
