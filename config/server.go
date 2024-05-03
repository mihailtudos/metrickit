package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/pkg/flags"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"

var addr *flags.ServerAddr

type EnvServerAddress struct {
	Address *string `env:"ADDRESS"`
}

func parseServerEnvs() {
	var envConfig EnvServerAddress
	if err := env.Parse(&envConfig); err != nil {
		log.Panic("failed to pars ADDRESS ENV")
	}

	host, port, err := splitAddressParts(envConfig.Address)
	if err != nil {
		addr = flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
		_ = flag.Value(addr)
		flag.Var(addr, "a", "server address - usage: ADDRESS:PORT")
		flag.Parse()
	} else {
		addr = flags.NewServerAddressFlag(host, port)
	}
}

type ServerConfig struct {
	Address string
	Log     *slog.Logger
}

func NewServerConfig() *ServerConfig {
	parseServerEnvs()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &ServerConfig{Address: addr.String(), Log: logger}
}

func splitAddressParts(address *string) (string, int, error) {
	if address == nil {
		return "", 0, fmt.Errorf("invalid address: missing parts")
	}
	parts := strings.Split(*address, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address: missing parts")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid address: port must be an int value")
	}

	return parts[0], port, nil
}
