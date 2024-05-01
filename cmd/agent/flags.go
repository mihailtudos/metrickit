package main

import (
	"flag"
	"github.com/mihailtudos/metrickit/pkg/flags"
	"time"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"
const DefaultReportInterval = 10
const DefaultPoolInterval = 2

var serverAddr *flags.ServerAddr
var poolIntervalInSeconds *flags.DurationFlag
var reportIntervalInSeconds *flags.DurationFlag

func parseFlags() {
	serverAddr = flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
	poolIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultPoolInterval)
	reportIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultReportInterval)

	_ = flag.Value(serverAddr)

	flag.Var(serverAddr, "a", "server address - usage: ADDRESS:PORT")
	flag.Var(poolIntervalInSeconds, "p", "frequency of polling the metrics")
	flag.Var(reportIntervalInSeconds, "r", "frequency of sending metrics to the server in seconds")

	flag.Parse()
}
