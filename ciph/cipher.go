package ciph

import (
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

func PickCipher(name string, password string) (Cipher, error) {
	name = strings.ToUpper(name)
	switch name {
	case "DUMMY":
		return &dummy{}, nil
	}
	return &dummy{}, nil
}

type dummy struct{}

func (dummy) StreamConn(c net.Conn) net.Conn             { return c }
func (dummy) PacketConn(c net.PacketConn) net.PacketConn { return c }
