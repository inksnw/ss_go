package main

import (
	"flag"
	"fmt"
	"github.com/ss_go/socks"
	"github.com/ss_go/core"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var config struct{
	Verbose bool
	UDPTimeout time.Duration
}

func main()  {
	var flags struct{
		Client string
		Server string
		Cipher string
		Password string
		Socks string
		UDPSocks  bool
	}
	flag.BoolVar(&config.Verbose,"verbose",false,"详细日志模式")
	flag.StringVar(&flags.Cipher, "cipher", "AEAD_CHACHA20_POLY1305", "加密方式")
	flag.StringVar(&flags.Server,"s","","服务端地址")
	flag.StringVar(&flags.Client, "c", "", "客户端地址")
	flag.StringVar(&flags.Password, "password", "", "密码")
	flag.StringVar(&flags.Socks, "socks", "", "客户端监听地址")
	flag.BoolVar(&flags.UDPSocks, "u", false, "(client-only) Enable UDP support for SOCKS")
	flag.Parse()

	var key []byte
	if flags.Client !=""{
		addr :=flags.Client
		cipher :=flags.Cipher
		password := flags.Password
		var err error

		if strings.HasPrefix(addr,"ss://"){
			addr,cipher,password,err=parseURL(addr)
			if err!=nil{
				log.Fatal(err)
			}

		}
		ciph, err := core.PickCipher(cipher, key, password)
		if err!=nil{
			log.Fatal(err)
		}

		if flags.Socks != ""{
			socks.UDPEnabled=flags.UDPSocks

			go socksLocal(flags.Socks,addr,ciph.StreamConn)

		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

	}

}

func parseURL(s string) (addr,cipher,password string,err error) {
	u,err:=url.Parse(s)
	if err!=nil{
		return
	}
	addr=u.Host
	if u.User!=nil{
		cipher=u.User.Username()
		password,_=u.User.Password()
	}
	return
}
var logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)
func logf(f string, v ...interface{}) {
	if config.Verbose {
		logger.Output(2, fmt.Sprintf(f, v...))
	}
}