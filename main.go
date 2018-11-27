package main

import (
	"flag"
	"github.com/ss_go/core"
	"github.com/ss_go/socks"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var config struct {
	UDPTimeout time.Duration
}

var flags struct {
	Client   string
	Server   string
	Socks    string
	UDPSocks bool
	UDPTun   string
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
	flag.BoolVar(&flags.UDPSocks, "u", false, "(client-only) Enable UDP support for SOCKS")
	flag.StringVar(&flags.UDPTun, "udptun", "", "(client-only) UDP tunnel (laddr1=raddr1,laddr2=raddr2,...)")
	flag.Parse()
	if flags.Client == "" && flags.Server == "" {
		flag.Usage()
		return
	}
}

func client() {
	addr, cipher, password := parseURL(flags.Client)
	ciph := core.PickCipher(cipher, password)

	go socksLocal(flags.Socks, addr, ciph.StreamConn)

	if flags.UDPTun != "" {
		for _, tun := range strings.Split(flags.UDPTun, ",") {
			p := strings.Split(tun, "=")
			go udpLocal(p[0], addr, p[1], ciph.PacketConn)
		}
	}

	socks.UDPEnabled = flags.UDPSocks
	if flags.UDPSocks {
		go udpSocksLocal(flags.Socks, addr, ciph.PacketConn)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

}

func server() {
	addr, cipher, password := parseURL(flags.Server)
	ciph := core.PickCipher(cipher, password)
	go tcpRemote(addr, ciph.StreamConn)
	go udpRemote(addr, ciph.PacketConn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(addrString string) (addr, cipher, password string) {
	u, err := url.Parse(addrString)
	if err != nil {
		panic(err)
	}
	addr = u.Host
	cipher = u.User.Username()
	password, _ = u.User.Password()
	return
}
