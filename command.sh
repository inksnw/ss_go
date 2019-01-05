#!/usr/bin/env bash
go run *.go -c 'ss://AEAD_CHACHA20_POLY1305:your-password@127.0.0.1:8488'  -socks :1081
go run *.go -s 'ss://AEAD_CHACHA20_POLY1305:your-password@:8488' 


