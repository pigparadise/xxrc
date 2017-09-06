package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [config]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var listen_addr string
	flag.StringVar(&listen_addr, "listen_addr", "0.0.0.0:10000", "listen addr(0.0.0.0:10000)")
	flag.Usage = usage
	flag.Parse()

	logic_init()

	log.Printf("start http service<%s>", listen_addr)
	log.Fatal(http.ListenAndServe(listen_addr, nil))
}
