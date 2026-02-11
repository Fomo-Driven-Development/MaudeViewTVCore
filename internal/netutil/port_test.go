package netutil

import (
	"net"
	"testing"
)

func TestSelectBindAddrPreferredFree(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
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
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen busy: %v", err)
	}
	defer func() { _ = busy.Close() }()

	free, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen free: %v", err)
	}
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
