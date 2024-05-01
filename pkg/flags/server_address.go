package flags

import (
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
	parts := strings.Split(flagsValue, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid server addres format usage: ADDRESS:PORT (e.g: localhost:8080)")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	sa.Port = port
	sa.Host = parts[0]
	return nil
}
