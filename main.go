package main

import (
	"flag"
	"github.com/ss_go/socks"
	"log"
	"net"
)

var flags struct {
	Type string
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&flags.Type, "t", "", "c/s")
	flag.Parse()
	if flags.Type == "c" {
		client()
	} else if flags.Type == "s" {
		server()
	}
}

func server() {
	listener, err := net.Listen("tcp", ":8787")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go socks.TcpRemote(conn)
	}

}

func client() {
	listener, err := net.Listen("tcp", ":8889")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go socks.TcpLocal(conn, "127.0.0.1:8787")

	}

}
