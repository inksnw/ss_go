package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ss_go/core"
)

var config struct {
	Verbose bool
}

func main() {
	var flags struct {
		Client   string
		Server   string
		Cipher   string
		Password string
		Socks    string
		UDPSocks bool
	}
	flag.BoolVar(&config.Verbose, "verbose", false, "详细日志模式")
	flag.StringVar(&flags.Cipher, "cipher", "AEAD_CHACHA20_POLY1305", "加密方式")
	flag.StringVar(&flags.Server, "s", "", "服务端地址")
	flag.StringVar(&flags.Client, "c", "", "客户端地址")
	flag.StringVar(&flags.Password, "password", "", "密码")
	flag.StringVar(&flags.Socks, "socks", "", "客户端监听地址")
	flag.Parse()

	if flags.Client != "" {
		addr := flags.Client
		cipher := flags.Cipher
		password := flags.Password
		var err error
		if strings.HasPrefix(addr, "ss://") {
			addr, cipher, password, err = parseURL(addr)
			CheckErr(err)
		}
		ciph, err := core.PickCipher(cipher, password)
		CheckErr(err)

		if flags.Socks != "" {
			go socksLocal(flags.Socks, addr, ciph.StreamConn)
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

	}

	if flags.Server != "" { // server mode
		addr := flags.Server
		cipher := flags.Cipher
		password := flags.Password
		var err error
		if strings.HasPrefix(addr, "ss://") {
			addr, cipher, password, err = parseURL(addr)
			if err != nil {
				log.Fatal(err)
			}
		}

		ciph, err := core.PickCipher(cipher, password)
		if err != nil {
			log.Fatal(err)
		}
		go tcpRemote(addr, ciph.StreamConn)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}

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
