package main

import (
	"flag"
	"github.com/mihailtudos/metrickit/pkg/flags"
	"time"
)

const DEFAULT_PORT = 8080
const DEFAULT_ADDRESS = "localhost"
const DEFAULT_REPORT_INTERVAL = 10
const DEFAULT_POLL_INTERVAL = 2

var serverAddr *flags.ServerAddr
var poolIntervalInSeconds *flags.DurationFlag
var reportIntervalInSeconds *flags.DurationFlag

func parseFlags() {
	serverAddr = flags.NewServerAddressFlag(DEFAULT_ADDRESS, DEFAULT_PORT)
	poolIntervalInSeconds = flags.NewDurationFlag(time.Second, DEFAULT_POLL_INTERVAL)
	reportIntervalInSeconds = flags.NewDurationFlag(time.Second, DEFAULT_REPORT_INTERVAL)

	_ = flag.Value(serverAddr)

	flag.Var(serverAddr, "a", "server address - usage: ADDRESS:PORT")
	flag.Var(poolIntervalInSeconds, "p", "frequency of polling the metrics")
	flag.Var(reportIntervalInSeconds, "r", "frequency of sending metrics to the server in seconds")

	flag.Parse()
}
