package socks

import (
	"log"
	"net"
)

func TcpLocal(localConn net.Conn, server string) {
	defer localConn.Close()
	targetAddr, err := HandelShake(localConn)
	if err != nil {
		//log.Print()
	}

	if err != nil {
		log.Print(err)
	}
	log.Print("请求完整地址", targetAddr)

	remoteConn, err := net.Dial("tcp", server)
	remoteConn.Write(targetAddr)

	if err != nil {
		log.Print(err)
		localConn.Close()
		return
	}
	Relay(remoteConn, localConn)

}

func TcpRemote(conn net.Conn) {
	defer conn.Close()

	addr, err := ReadAddr(conn)
	if err != nil {
		log.Print(err)
	}

	var remote net.Conn
	remote, err = net.Dial("tcp", addr.String())

	if err != nil {
		log.Print(err)
		conn.Close()
		return
	}
	log.Print("请求完整地址", addr.String())

	Relay(conn, remote)
}
