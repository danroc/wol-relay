package main

import (
	"bytes"
	"net"
	"testing"
)

var testNetworks = []net.IPNet{
	{IP: net.IPv4(192, 168, 1, 1), Mask: net.CIDRMask(24, 32)},
	{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)},
}

func TestContainsIP(t *testing.T) {
	ips := []net.IP{
		net.IPv4(192, 168, 1, 5),
		net.IPv4(10, 0, 0, 5),
	}
	tests := []struct {
		ip       net.IP
		expected bool
	}{
		{net.IPv4(192, 168, 1, 5), true},
		{net.IPv4(10, 0, 0, 5), true},
		{net.IPv4(192, 168, 1, 6), false},
		{net.IPv4(10, 0, 0, 4), false},
		{net.IPv4(8, 8, 8, 8), false},
		// IPv4-mapped IPv6 form of 192.168.1.5; must compare equal.
		{net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 168, 1, 5}, true},
	}

	for _, tt := range tests {
		got := containsIP(ips, tt.ip)
		if got != tt.expected {
			t.Errorf(
				"containsIP(ips, %v) = %v; want %v",
				tt.ip, got, tt.expected,
			)
		}
	}
}

func TestLocalIPs(t *testing.T) {
	ips, err := localIPs()
	if err != nil {
		t.Fatalf("localIPs() unexpected error: %v", err)
	}
	if len(ips) == 0 {
		t.Fatal("localIPs() returned no addresses; expected at least loopback")
	}
	foundLoopback := false
	for _, ip := range ips {
		if ip.To4() == nil {
			t.Errorf("localIPs() returned non-IPv4 address: %v", ip)
		}
		if ip.IsLoopback() {
			foundLoopback = true
		}
	}
	if !foundLoopback {
		t.Errorf("localIPs() did not include loopback; got %v", ips)
	}
}

func TestIsIPInAny(t *testing.T) {
	tests := []struct {
		ip       net.IP
		expected bool
	}{
		{net.IPv4(192, 168, 1, 0), true},
		{net.IPv4(192, 168, 1, 1), true},
		{net.IPv4(192, 168, 1, 100), true},
		{net.IPv4(192, 168, 1, 255), true},
		{net.IPv4(192, 168, 0, 0), false},
		{net.IPv4(10, 0, 0, 0), true},
		{net.IPv4(10, 0, 0, 1), true},
		{net.IPv4(10, 0, 0, 100), true},
		{net.IPv4(10, 0, 0, 255), true},
		{net.IPv4(10, 0, 1, 0), false},
		{net.IPv4(172, 16, 0, 1), false},
		{net.IPv4(0, 0, 0, 0), false},
	}

	for _, tt := range tests {
		got := isIPInAny(tt.ip, testNetworks)
		if got != tt.expected {
			t.Errorf(
				"isIPInAny(%v, networks) = %v; want %v",
				tt.ip, got, tt.expected,
			)
		}
	}
}

func TestToBroadcastIP(t *testing.T) {
	tests := []struct {
		network  net.IPNet
		expected net.IP
		isValid  bool
	}{
		{
			net.IPNet{
				IP:   net.IPv4(192, 168, 1, 1),
				Mask: net.CIDRMask(24, 32),
			},
			net.IPv4(192, 168, 1, 255),
			true,
		},
		{
			net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.CIDRMask(8, 32),
			},
			net.IPv4(10, 255, 255, 255),
			true,
		},
		{
			net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.CIDRMask(8, 128), // Invalid mask length
			},
			nil,
			false,
		},
		{
			net.IPNet{
				IP:   net.IPv4(192, 168, 2, 5),
				Mask: net.CIDRMask(32, 32),
			},
			net.IPv4(192, 168, 2, 5),
			true,
		},
	}

	for _, tt := range tests {
		got, err := toBroadcastIP(tt.network)
		if err != nil && tt.isValid {
			t.Errorf("toBroadcastIP(%q) unexpected error: %v", tt.network, err)
			continue
		}
		if err == nil && !tt.isValid {
			t.Errorf("toBroadcastIP(%q) expected error, got nil", tt.network)
			continue
		}
		if !got.Equal(tt.expected) {
			t.Errorf(
				"toBroadcastIP(%q) = %v, want %v",
				tt.network, got, tt.expected,
			)
		}
	}
}

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		input    string
		expected net.IPNet
		isValid  bool
	}{
		{"10.0.0.0/24", net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(24, 32)}, true},
		{"10.0.0.5/24", net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(24, 32)}, true},
		{"192.168.2.5/32", net.IPNet{IP: net.IPv4(192, 168, 2, 5), Mask: net.CIDRMask(32, 32)}, true},
		{"172.16.0.0/16", net.IPNet{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(16, 32)}, true},
		{"not-a-cidr", net.IPNet{}, false},
		{"::1/128", net.IPNet{}, false},
		{"10.0.0.1/129", net.IPNet{}, false},
	}

	for _, tt := range tests {
		got, err := parseCIDR(tt.input)
		if err != nil && tt.isValid {
			t.Errorf("parseCIDR(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if err == nil && !tt.isValid {
			t.Errorf("parseCIDR(%q) expected error, got nil", tt.input)
			continue
		}
		if tt.isValid && (!got.IP.Equal(tt.expected.IP) || !bytes.Equal(got.Mask, tt.expected.Mask)) {
			t.Errorf("parseCIDR(%q) IP = %v, want %v; Mask = %v, want %v", tt.input, got.IP, tt.expected.IP, got.Mask, tt.expected.Mask)
		}
	}
}
