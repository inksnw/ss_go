package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ss_go/core"
)

var config struct {
	Verbose    bool
	UDPTimeout time.Duration
}
var flags struct {
	Client   string
	Server   string
	Cipher   string
	Password string
	Socks    string
}

func main() {
	init_flag()
	if flags.Client != "" {
		client()
	} else if flags.Server != "" {
		server()
	}

}

func init_flag() {
	flag.BoolVar(&config.Verbose, "verbose", false, "详细日志")
	flag.StringVar(&flags.Server, "s", "", "服务端")
	flag.StringVar(&flags.Client, "c", "", "客户端")
	flag.StringVar(&flags.Socks, "socks", "", "客户端监听")
	flag.Parse()
	if flags.Client == "" && flags.Server == "" {
		flag.Usage()
		return
	}
}

func client() {
	addrString := flags.Client
	addr, cipher, password, _ := parseURL(addrString)
	ciph, _ := core.PickCipher(cipher, password)
	go socksLocal(flags.Socks, addr, ciph.StreamConn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

}

func server() {
	addrString := flags.Server
	addr, cipher, password, _ := parseURL(addrString)
	ciph, _ := core.PickCipher(cipher, password)
	go tcpRemote(addr, ciph.StreamConn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(addrString string) (addr, cipher, password string, err error) {
	u, err := url.Parse(addrString)
	CheckErr(err)
	addr = u.Host
	cipher = u.User.Username()
	password, _ = u.User.Password()
	return
}

var logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

func logf(f string, v ...interface{}) {
	if config.Verbose {
		logger.Output(2, fmt.Sprintf(f, v...))
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
