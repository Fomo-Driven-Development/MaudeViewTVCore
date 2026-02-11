package netutil

import (
	"errors"
	"fmt"
	"net"
)

// SelectBindAddr picks an available bind address based on preferred and fallback list.
func SelectBindAddr(preferred string, candidates []string, autoFallback bool) (string, error) {
	if preferred != "" {
		ok, err := IsAddrAvailable(preferred)
		if err != nil {
			return "", err
		}
		if ok {
			return preferred, nil
		}
		if !autoFallback {
			return "", fmt.Errorf("preferred bind address in use: %s", preferred)
		}
	}

	for _, addr := range candidates {
		ok, err := IsAddrAvailable(addr)
		if err != nil {
			return "", err
		}
		if ok {
			return addr, nil
		}
	}

	return "", errors.New("no available controller bind addresses")
}

// IsAddrAvailable returns true when an address can be listened on.
func IsAddrAvailable(addr string) (bool, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false, nil
	}
	if closeErr := ln.Close(); closeErr != nil {
		return false, closeErr
	}
	return true, nil
}
