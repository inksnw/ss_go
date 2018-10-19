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
	Verbose bool
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
	flag.BoolVar(&config.Verbose, "verbose", false, "详细日志模式")
	flag.StringVar(&flags.Server, "s", "", "服务端地址")
	flag.StringVar(&flags.Client, "c", "", "客户端地址")
	flag.StringVar(&flags.Socks, "socks", "", "客户端监听地址")
	flag.Parse()
	if flags.Client == "" && flags.Server == "" {
		flag.Usage()
		return
	}
}



func client() {
	addr := flags.Client
	cipher := flags.Cipher
	password := flags.Password
	addr, cipher, password, _ = parseURL(addr)
	ciph, _ := core.PickCipher(cipher, password)
	go socksLocal(flags.Socks, addr, ciph.StreamConn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

}

func server() {
	addr := flags.Server
	cipher := flags.Cipher
	password := flags.Password
	addr, cipher, password, _ = parseURL(addr)
	ciph, _ := core.PickCipher(cipher, password)
	go tcpRemote(addr, ciph.StreamConn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(s string) (addr, cipher, password string, err error) {
	u, err := url.Parse(s)
	CheckErr(err)
	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
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
