package fetcher

import (
	"errors"
	"net"
	"net/http"
	"syscall"
	"time"
)

// blockedCIDRs holds non-public ranges that net.IP's helper methods don't catch.
var blockedCIDRs = func() []*net.IPNet {
	raw := []string{
		"0.0.0.0/8",       // "this network" (RFC 1122)
		"100.64.0.0/10",   // carrier-grade NAT (RFC 6598)
		"192.0.0.0/24",    // IETF protocol assignments
		"192.0.2.0/24",    // TEST-NET-1 (documentation)
		"198.18.0.0/15",   // benchmarking (RFC 2544)
		"198.51.100.0/22", // TEST-NET-2 (documentation) - FIXED from /24 to /22
		"203.0.113.0/24",  // TEST-NET-3 (documentation)
		"240.0.0.0/4",     // reserved (former class E)

		// IPv6 Ranges
		"64:ff9b::/96",  // NAT64
		"100::/64",      // Discard-Only prefix (RFC 6666)
		"2001:10::/28",  // ORCHIDv2 (RFC 7343)
		"2001:db8::/32", // IPv6 documentation
	}

	var nets []*net.IPNet
	for _, c := range raw {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic("fetcher: invalid CIDR " + c)
		}
		nets = append(nets, n)
	}
	return nets
}()

// isPublicIP reports whether ip is safe to connect to
// i.e. it is NOT loopsback, private, link-local, or unspecified
func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Normalize IPv4-mapped to IPv6 (like ::ffff:127.0.0.1) down to plain IPv6
	// Prevents smuggling a private IPv4 through IPv6 wrapper and dodge IPv4 checks
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}

	if ip.IsUnspecified() || // 0.0.0.0, ::
		ip.IsLoopback() || // 127.0.0.0/8, ::1
		ip.IsPrivate() || // 10/8, 172.16/12, 192.168/16, fc00::/7
		ip.IsLinkLocalUnicast() || // 169.254/16, fe80::/10
		ip.IsMulticast() { // 224/4, ff00::/8 (covers link-local & interface-local multicast)
		return false
	}

	for _, blocked := range blockedCIDRs {
		if blocked.Contains(ip) {
			return false
		}
	}

	return true
}

// safeClient creates an http.Client that is immune to SSRF and DNS Rebinding
func safeClient() *http.Client {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			// address looks like "93.184.216.34:443" or "[2001:db8::1]:80"
			// We need to split the host (IP) form the port
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}

			// Turn the string into a real net.IP object
			ip := net.ParseIP(host)
			if ip == nil {
				return errors.New("safeClient: invalid IP address in dialer")
			}

			// Run the bounder function
			if !isPublicIP(ip) {
				return errors.New("safeClient: blocked attempt to connect to private/internal IP")
			}

			// At this point the bouncer approved it
			return nil
		},
	}

	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

	// Build and return the Boss (Client) and the Manager (Transport)
	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
}
