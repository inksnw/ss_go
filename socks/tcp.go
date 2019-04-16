package socks

import (
	"log"
	"net"
)

func TcpLocal(localConn net.Conn, server string) {
	defer localConn.Close()
	//_ = localConn.(*net.TCPConn).SetKeepAlive(true)
	targetAddr, err := handShake(localConn)
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
