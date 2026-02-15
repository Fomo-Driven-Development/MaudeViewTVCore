package netutil

import (
	"net"
	"strings"
	"testing"
)

func mustListen(t *testing.T, network, address string) net.Listener {
	t.Helper()
	ln, err := net.Listen(network, address)
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "permission denied") {
			t.Skipf("skipping network bind test in restricted environment: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	return ln
}

func TestSelectBindAddrPreferredFree(t *testing.T) {
	ln := mustListen(t, "tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()

	got, err := SelectBindAddr(addr, nil, false)
	if err != nil {
		t.Fatalf("SelectBindAddr() error = %v", err)
	}
	if got != addr {
		t.Fatalf("SelectBindAddr() = %q, want %q", got, addr)
	}
}

func TestSelectBindAddrFallback(t *testing.T) {
	busy := mustListen(t, "tcp", "127.0.0.1:0")
	defer func() { _ = busy.Close() }()

	free := mustListen(t, "tcp", "127.0.0.1:0")
	freeAddr := free.Addr().String()
	_ = free.Close()

	got, err := SelectBindAddr(busy.Addr().String(), []string{busy.Addr().String(), freeAddr}, true)
	if err != nil {
		t.Fatalf("SelectBindAddr() error = %v", err)
	}
	if got != freeAddr {
		t.Fatalf("SelectBindAddr() = %q, want %q", got, freeAddr)
	}
}
