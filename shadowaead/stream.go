package shadowaead

import (
	"crypto/cipher"
	"io"
	"net"
)

type streamConn struct {
	net.Conn
	Cipher
	r *reader
	w *writer
}
type reader struct {
	io.Reader
	cipher.AEAD
	nonce    []byte
	buf      []byte
	leftover []byte
}
type writer struct {
	io.Writer
	cipher.AEAD
	nonce []byte
	buf   []byte
}

func NewConn(c net.Conn, ciph Cipher) net.Conn {
	return &streamConn{Conn: c, Cipher: ciph}

}
