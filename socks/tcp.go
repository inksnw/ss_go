package socks

import (
	"io"
	"log"
	"net"
)

func SocksLocal(laddr, server string, shadow func(net.Conn) net.Conn) {
	log.Printf("use socks proxy")
	tcpLocal(laddr, server, handShakeGetAddr, shadow)
}

func TcpTun(addr, server, target string, shadow func(net.Conn) net.Conn) {
	log.Printf("use tcp tun")

	tgt := ParseAddr(target)
	if tgt == nil {
		log.Printf("invalid target address %q", target)
		return
	}
	log.Printf("TCP tunnel %s <-> %s <-> %s", addr, server, target)
	getaddr := func(io.ReadWriter) (Addr, error) { return tgt, nil }
	tcpLocal(addr, server, getaddr, shadow)

}

func tcpLocal(addr, server string, handShakeGetAddr func(conn io.ReadWriter) (addr Addr, err error), shadow func(net.Conn) net.Conn) {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {

			defer conn.Close()
			//_ = conn.(*net.TCPConn).SetKeepAlive(true)
			//targetAddr, err := handShake(conn)
			targetAddr, err := handShakeGetAddr(conn)
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
			remoteConn = shadow(remoteConn)

			Relay(remoteConn, conn)
		}()

	}

}

func TcpRemote(addr string, shadow func(net.Conn) net.Conn) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer conn.Close()
			_ = conn.(*net.TCPConn).SetKeepAlive(true)
			conn = shadow(conn)
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

		}()
	}

}
