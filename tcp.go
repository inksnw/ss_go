package main

import (
	"fmt"
	"github.com/ss_go/socks"
	"io"
	"net"
	"time"
)

// Create a SOCKS server listening on addr and proxy to server.
func socksLocal(addr, server string) {

	getAddr := func(c net.Conn) (socks.Addr, error) {
		return socks.Handshake(c)
	}

	tcpLocal(addr, server, getAddr)

}
func tcpLocal(addr, server string, getAddr func(net.Conn) (socks.Addr, error)) {
	localServer, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("failed to listen on %s: %v", addr, err)
		return
	} else {
		fmt.Printf("本机->远程主机  %s -> %s\n", addr, server)
	}

	for {
		localConn, err := localServer.Accept()

		if err != nil {
			fmt.Printf("failed to accept: %s", err)
			continue
		}
		go func() {
			defer localConn.Close()
			localConn.(*net.TCPConn).SetKeepAlive(true)
			targetAddr, err := getAddr(localConn)
			if err != nil {
				fmt.Printf("failed to get target address: %v", err)
				return
			}
			remoteConn, err := net.Dial("tcp", server)
			if err != nil {
				fmt.Printf("failed to connect to server %v: %v", server, err)
				return
			}
			defer remoteConn.Close()
			remoteConn.(*net.TCPConn).SetKeepAlive(true)

			if _, err = remoteConn.Write(targetAddr); err != nil {
				fmt.Printf("failed to send target address: %v", err)
				return
			}
			_, _, err = relay(remoteConn, localConn)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				fmt.Printf("relay error: %v", err)
			}
		}()

	}
}

// Listen on addr for incoming connections.
func tcpRemote(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("failed to listen on %s: %v", addr, err)
		return
	}

	fmt.Printf("listening TCP on %s", addr)
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("failed to accept: %v", err)
			continue
		}

		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)

			tgt, err := socks.ReadAddr(c)
			if err != nil {
				fmt.Printf("failed to get target address: %v", err)
				return
			}

			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				fmt.Printf("failed to connect to target: %v", err)
				return
			}
			defer rc.Close()
			rc.(*net.TCPConn).SetKeepAlive(true)

			fmt.Printf("proxy %s <-> %s", c.RemoteAddr(), tgt)
			_, _, err = relay(c, rc)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				fmt.Printf("relay error: %v", err)
			}
		}()
	}
}

func relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)
	go func() {
		n, err := io.Copy(right, left)
		right.SetDeadline(time.Now())
		left.SetDeadline(time.Now())
		ch <- res{n, err}
	}()

	n, err := io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}
	return n, rs.N, err

}
