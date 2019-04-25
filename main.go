package main

import (
	"flag"
	"github.com/ss_go/socks"
	"log"
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
	socks.TcpRemote(":8787")
}

func client() {
	socks.SocksLocal(":8889", "127.0.0.1:8787")
}
