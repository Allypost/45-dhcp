package ports

import (
	"fmt"
	"strconv"
)

func ParsePort(ports string) (int16, error) {
	port, err := strconv.ParseInt(ports, 10, 16)

	if err != nil {
		return 0, err
	}

	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("invalid port: %d", port)
	}

	return int16(port), nil
}
