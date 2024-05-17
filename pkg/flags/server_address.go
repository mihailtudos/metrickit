package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ServerAddr struct {
	Host string
	Port int
}

func NewServerAddressFlag(address string, port int) *ServerAddr {
	srvAddr := new(ServerAddr)
	srvAddr.Port = port
	srvAddr.Host = address
	return srvAddr
}

func (sa *ServerAddr) String() string {
	return fmt.Sprintf("%s:%d", sa.Host, sa.Port)
}

func (sa *ServerAddr) Set(flagsValue string) error {
	const addressPortLength = 2
	parts := strings.Split(flagsValue, ":")
	if len(parts) != addressPortLength {
		return errors.New("invalid server addres format usage: ADDRESS:PORT (e.g: localhost:8080)")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("failed to convert port string '%s' to integer: %w", parts[1], err)
	}

	sa.Port = port
	sa.Host = parts[0]
	return nil
}
