package main

import (
	"flag"
	"log"
	"net/http"
)

type Options struct {
	ListenAddr string
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	var options Options
	flag.StringVar(&options.ListenAddr, "listen_addr", "0.0.0.0:10003", "listen port(0.0.0.0:10003)")

	logic_init()

	addr := options.ListenAddr
	log.Printf("start http service<%s>", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
