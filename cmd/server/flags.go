package main

import (
	"flag"
	"github.com/mihailtudos/metrickit/pkg/flags"
)

const DEFAULT_PORT = 8080
const DEFAULT_ADDRESS = "localhost"

var addr *flags.ServerAddr

func parseFlags() {

	addr = flags.NewServerAddressFlag(DEFAULT_ADDRESS, DEFAULT_PORT)
	_ = flag.Value(addr)
	flag.Var(addr, "a", "server address - usage: ADDRESS:PORT")
	flag.Parse()
}
