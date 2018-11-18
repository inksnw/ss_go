#!/usr/bin/env bash
go run *.go -c 'ss://AEAD_CHACHA20_POLY1305:your-password@[0.0.0.0]:8488'  -socks :1081
go run *.go -c 'ss://AEAD_CHACHA20_POLY1305:your-password@[192.168.25.131]:8488'  -socks :1081
go run *.go -s 'ss://AEAD_CHACHA20_POLY1305:your-password@:8488'