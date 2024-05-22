package config

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/pkg/flags"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"

var addr *flags.ServerAddr

type EnvServerAddress struct {
	Address *string `env:"ADDRESS"`
}

func parseServerEnvs() error {
	var envConfig EnvServerAddress
	if err := env.Parse(&envConfig); err != nil {
		return errors.New("failed to pars ADDRESS ENV: " + err.Error())
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

	return nil
}

type ServerConfig struct {
	Log     *slog.Logger
	Address string
}

func NewServerConfig() (*ServerConfig, error) {
	err := parseServerEnvs()
	if err != nil {
		return nil, errors.New("failed to create the config " + err.Error())
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &ServerConfig{Address: addr.String(), Log: logger}, nil
}

func splitAddressParts(address *string) (string, int, error) {
	const numberOfHostPortParts = 2
	if address == nil {
		return "", 0, errors.New("invalid address: missing parts")
	}
	parts := strings.Split(*address, ":")
	if len(parts) != numberOfHostPortParts {
		return "", 0, errors.New("invalid address: missing parts")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, errors.New("invalid address: port must be an int value")
	}

	return parts[0], port, nil
}
