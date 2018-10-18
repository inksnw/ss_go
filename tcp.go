package main

import (
	"github.com/ss_go/socks"
	"io"
	"net"
	"time"
)

// Create a SOCKS server listening on addr and proxy to server.
func socksLocal(addr, server string, shadow func(net.Conn) net.Conn) {
	tcpLocal(addr, server, shadow, func(c net.Conn) (socks.Addr, error) { return socks.Handshake(c) })

}
func tcpLocal(addr, server string, shadow func(net.Conn) net.Conn, getAddr func(net.Conn) (socks.Addr, error)) {
	local_server, err := net.Listen("tcp", addr)
	if err != nil {
		logf("failed to listen on %s: %v", addr, err)
		return
	} else {
		logf("listen on %s", addr)
		logf("SOCKS proxy %s <-> %s", addr, server)
	}

	for {
		local_conn, err := local_server.Accept()
		if err != nil {
			logf("failed to accept: %s", err)
			continue
		}
		go func() {
			defer local_conn.Close()
			local_conn.(*net.TCPConn).SetKeepAlive(true)
			target_addr, err := getAddr(local_conn)
			if err != nil {
				// UDP: keep the connection until disconnect then free the UDP socket
				if err == socks.InfoUDPAssociate {
					buf := []byte{}
					// block here
					for {
						_, err := local_conn.Read(buf)
						if err, ok := err.(net.Error); ok && err.Timeout() {
							continue
						}
						logf("UDP Associate End.")
						return
					}
				}

				logf("failed to get target address: %v", err)
				return
			}

			remote_conn, err := net.Dial("tcp", server)
			if err != nil {
				logf("failed to connect to server %v: %v", server, err)
				return
			}
			defer remote_conn.Close()
			remote_conn.(*net.TCPConn).SetKeepAlive(true)
			remote_conn = shadow(remote_conn)
			if _, err = remote_conn.Write(target_addr); err != nil {
				logf("failed to send target address: %v", err)
				return
			}
			_, _, err = relay(remote_conn, local_conn)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				logf("relay error: %v", err)
			}
		}()

	}
}

// Listen on addr for incoming connections.
func tcpRemote(addr string, shadow func(net.Conn) net.Conn) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		logf("failed to listen on %s: %v", addr, err)
		return
	}

	logf("listening TCP on %s", addr)
	for {
		c, err := l.Accept()
		if err != nil {
			logf("failed to accept: %v", err)
			continue
		}

		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)
			c = shadow(c)

			tgt, err := socks.ReadAddr(c)
			if err != nil {
				logf("failed to get target address: %v", err)
				return
			}

			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				logf("failed to connect to target: %v", err)
				return
			}
			defer rc.Close()
			rc.(*net.TCPConn).SetKeepAlive(true)

			logf("proxy %s <-> %s", c.RemoteAddr(), tgt)
			_, _, err = relay(c, rc)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				logf("relay error: %v", err)
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
