package main

import (
	"flag"
	"github.com/mihailtudos/metrickit/pkg/flags"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"

var addr *flags.ServerAddr

func parseFlags() {

	addr = flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
	_ = flag.Value(addr)
	flag.Var(addr, "a", "server address - usage: ADDRESS:PORT")
	flag.Parse()
}
