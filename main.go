package main

import (
	"flag"
	"github.com/ss_go/ciph"
	"github.com/ss_go/socks"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var flags struct {
	Type   string
	TCPTun string
	UDPTun string
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&flags.Type, "t", "", "c/s")
	flag.StringVar(&flags.TCPTun, "tcptun", "", "(client-only) TCP tunnel (laddr1=raddr1,laddr2=raddr2,...)")
	flag.StringVar(&flags.UDPTun, "ucptun", "", "(client-only) UDP tunnel (laddr1=raddr1,laddr2=raddr2,...)")
	flag.Parse()
	local := ":8889"
	server := "127.0.0.1:8787"
	serverSelf := ":8787"
	if flags.Type == "c" {

		ciph, err := ciph.PickCipher("DUMMY", "123")

		if err != nil {
			panic("choice ciph fail")
		}

		go socks.SocksLocal(local, server, ciph.StreamConn)
		if flags.TCPTun != "" {
			for _, tun := range strings.Split(flags.TCPTun, ",") {
				p := strings.Split(tun, "=")
				go socks.TcpTun(p[0], server, p[1], ciph.StreamConn)
			}
		}
		if flags.UDPTun != "" {
			for _, tun := range strings.Split(flags.UDPTun, ",") {
				p := strings.Split(tun, "=")
				go socks.UdpLocal(p[0], server, p[1], ciph.PacketConn)
			}
		}
	} else if flags.Type == "s" {
		go socks.TcpRemote(serverSelf)
		go socks.UdpRemote(serverSelf)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
