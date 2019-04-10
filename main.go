package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

func main() {
	listener, err := net.Listen("tcp", ":8889")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handelConn(conn)

	}

}
func readAddr(r *bufio.Reader) (string, error) {
	version, _ := r.ReadByte()
	if version != 5 {
		return "", errors.New("非socks5协议")
	}
	cmd, _ := r.ReadByte()
	if cmd != 1 {
		return "", errors.New("客户端请求方法不为CONNECT")
	}
	/*
	  数字“1”：CONNECT ；
	  数字“2”：BIND ；
	  数字“3”：UDP ASSOCIATE；
	*/
	r.ReadByte() //RSV保留字跳过

	addrType, _ := r.ReadByte()
	if addrType != 3 {
		return "", errors.New("讲求地址不为域名")
	}
	addrLen, _ := r.ReadByte()
	addr := make([]byte, addrLen)
	io.ReadFull(r, addr)
	var port int16
	binary.Read(r, binary.BigEndian, &port)
	return fmt.Sprintf("%s:%d", addr, port), nil

}

func handelConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	err := handelShake(r, conn)
	if err != nil {
		//log.Print()
	}
	addr, err := readAddr(r)
	if err != nil {
		log.Print(err)
	}
	log.Print("请求完整地址", addr)
	resp := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, _ = conn.Write(resp)

	var remote net.Conn
	remote, err = net.Dial("tcp", addr)

	if err != nil {
		log.Print(err)
		conn.Close()
		return
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(remote, r)
		remote.Close()
	}()

	go func() {
		defer conn.Close()
		io.Copy(conn, remote)
		conn.Close()
	}()

	wg.Wait()

}

func handelShake(r *bufio.Reader, conn net.Conn) error {
	version, _ := r.ReadByte()
	if version != 5 {
		return errors.New("该协议不是socks5协议")
	}
	methodLen, _ := r.ReadByte()
	buf := make([]byte, methodLen)
	io.ReadFull(r, buf)
	_, err := conn.Write([]byte{5, 0})
	return err

}
