package fetcher

import (
	"net"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// genuinely public
		{"google dns", "8.8.8.8", true},
		{"cloudflare", "1.1.1.1", true},
		{"public ipv6", "2606:4700:4700::1111", true},

		// loopback
		{"ipv4 loopback", "127.0.0.1", false},
		{"ipv4 loopback high", "127.255.255.254", false},
		{"ipv6 loopback", "::1", false},

		// private
		{"10/8", "10.0.0.1", false},
		{"172.16/12", "172.16.5.4", false},
		{"192.168/16", "192.168.1.1", false},
		{"ipv6 ULA", "fc00::1", false},

		// link-local — 169.254.169.254 is the cloud metadata endpoint
		{"cloud metadata", "169.254.169.254", false},
		{"ipv6 link-local", "fe80::1", false},

		// unspecified
		{"ipv4 unspecified", "0.0.0.0", false},
		{"ipv6 unspecified", "::", false},

		// other non-public
		{"cgnat", "100.64.0.1", false},
		{"multicast", "224.0.0.1", false},
		{"reserved class E", "240.0.0.1", false},

		// the bypass attempts — IPv4-mapped IPv6
		{"mapped loopback", "::ffff:127.0.0.1", false},
		{"mapped private", "::ffff:10.0.0.1", false},
		{"mapped metadata", "::ffff:169.254.169.254", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse test IP %q in %s", tt.ip, tt.name)
			}
			if got := isPublicIP(ip); got != tt.want {
				t.Errorf("isPublicIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPublicIP_Nil(t *testing.T) {
	if isPublicIP(nil) {
		t.Error("isPublicIP(nil) = true, want false")
	}
}
