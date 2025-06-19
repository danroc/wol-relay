package main

import (
	"net"
	"testing"
)

func TestIsIPOneOf(t *testing.T) {
	networks := []net.IPNet{
		{IP: net.IPv4(192, 168, 1, 1), Mask: net.CIDRMask(24, 32)},
		{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)},
	}

	tests := []struct {
		ip       net.IP
		expected bool
	}{
		{net.IPv4(192, 168, 1, 1), true},
		{net.IPv4(192, 168, 1, 0), false},
		{net.IPv4(192, 168, 0, 1), false},
		{net.IPv4(10, 0, 0, 1), true},
		{net.IPv4(10, 0, 0, 0), false},
		{net.IPv4(10, 0, 1, 1), false},
		{net.IPv4(8, 8, 8, 8), false},
	}

	for _, tt := range tests {
		got := isIPOneOf(tt.ip, networks)
		if got != tt.expected {
			t.Errorf(
				"isIPOneOf(%v, networks) = %v; want %v",
				tt.ip, got, tt.expected,
			)
		}
	}
}

func TestIsIPInAny(t *testing.T) {
	networks := []net.IPNet{
		{IP: net.IPv4(192, 168, 1, 1), Mask: net.CIDRMask(24, 32)},
		{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(24, 32)},
	}

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
		got := isIPInAny(tt.ip, networks)
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
