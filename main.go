package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var flags struct {
	Client string
	Server string
	Socks  string
}

func main() {
	initFlag()
	if flags.Client != "" {
		client()
	} else if flags.Server != "" {
		server()
	}
}

func initFlag() {
	flag.StringVar(&flags.Server, "s", "", "服务端参数")
	flag.StringVar(&flags.Client, "c", "", "客户端参数")
	flag.StringVar(&flags.Socks, "socks", "", "客户端监听地址")
	flag.Parse()
	if flags.Client == "" && flags.Server == "" {
		flag.Usage()
		return
	}
}

func client() {
	addr, _, _, _ := parseURL(flags.Client)
	//本机地址,远端地址
	go socksLocal(flags.Socks, addr)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

}

func server() {
	addr, _, _, err := parseURL(flags.Server)
	if err != nil {
		panic(err)
	}
	go tcpRemote(addr)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(addrString string) (addr, cipher, password string, err error) {
	u, err := url.Parse(addrString)
	if err != nil {
		panic(err)
	}
	addr = u.Host
	cipher = u.User.Username()
	password, _ = u.User.Password()
	return
}
