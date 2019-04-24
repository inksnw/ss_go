package socks

import (
	"io"
	"log"
	"net"
)

func SocksLocal(localConn net.Conn, server string) {
	log.Printf("use socks proxy")
	tcpLocal(localConn, server, handShake)
}

//func tcpTun()  {
//
//}

func tcpLocal(localConn net.Conn, server string, getAddr func(conn io.ReadWriter) (addr Addr, err error)) {
	defer localConn.Close()
	//_ = localConn.(*net.TCPConn).SetKeepAlive(true)
	//targetAddr, err := handShake(localConn)
	targetAddr, err := getAddr(localConn)
	if err != nil {
		log.Printf("failed to get target address: %v", err)
		return
	}

	remoteConn, err := net.Dial("tcp", server)
	if err != nil {
		log.Printf("failed to connect to server %v: %v", server, err)
		return
	}
	defer remoteConn.Close()
	_ = remoteConn.(*net.TCPConn).SetKeepAlive(true)

	if _, err = remoteConn.Write(targetAddr); err != nil {
		log.Printf("failed to send target address: %v", err)
		return
	}
	Relay(remoteConn, localConn)

}

func TcpRemote(conn net.Conn) {
	defer conn.Close()
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	s := make([]byte, MaxAddrLen)
	addr, err := readAddr(conn, s)
	if err != nil {
		log.Printf("failed to get target address: %v", err)
	}

	remote, err := net.Dial("tcp", addr.String())
	if err != nil {
		log.Printf("failed to connect to target: %v", err)
		return
	}
	defer remote.Close()
	_ = remote.(*net.TCPConn).SetKeepAlive(true)

	Relay(conn, remote)
}
