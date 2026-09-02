package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

var (
	ErrSSRFBlocked = errors.New("ssrf protection: request blocked to private or restricted network")
)

// SafeHTTPClient creates an *http.Client that prevents SSRF attacks.
// It blocks connections to localhost, private IPv4/IPv6, link-local, and multicast addresses.
// It explicitly denies auto-following redirects to prevent bypasses.
func SafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// Do not resolve manually, rely on the dialer, but we validate resolved IPs post-dial OR pre-dial.
			// The safest way is to resolve first, check all IPs, then dial one.
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			for _, ip := range ips {
				if isRestrictedIP(ip) {
					return nil, fmt.Errorf("%w: %s resolves to restricted IP %s", ErrSSRFBlocked, host, ip.String())
				}
			}

			// We found safe IPs. Now let the actual dialer connect.
			// Note: There's a slight TOCTOU window here with DNS, but Go's HTTP client
			// handles standard DNS resolution safely for most threat models when we check post-dial.
			// Let's dial manually to the first safe IP to completely close TOCTOU.
			
			for _, ip := range ips {
				if !isRestrictedIP(ip) {
					safeAddr := net.JoinHostPort(ip.String(), port)
					conn, err := dialer.DialContext(ctx, network, safeAddr)
					if err == nil {
						return conn, nil
					}
				}
			}

			return nil, fmt.Errorf("failed to dial safe IPs for %s", host)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Do not follow redirects automatically.
			// A redirect might point to http://127.0.0.1/
			return http.ErrUseLastResponse
		},
	}

	return client
}

// isRestrictedIP returns true if the IP is internal, private, loopback, or otherwise dangerous for SSRF.
func isRestrictedIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return true // blocks loopback, multicast, link-local, unspecified
	}
	if ip.IsPrivate() {
		return true // blocks RFC1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) and IPv6 private
	}
	
	// Explicitly block AWS metadata (169.254.169.254) which is link-local but we want to be paranoid
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return false
}
