package metadata

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var allowedHosts = map[string]struct{}{
	"api.themoviedb.org":   {},
	"image.tmdb.org":       {},
	"api4.thetvdb.com":     {},
	"musicbrainz.org":      {},
	"coverartarchive.org":  {},
	"webservice.fanart.tv": {},
}

// SafeHTTPClient returns an HTTP client with SSRF protections.
func SafeHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return validateOutboundURL(req.URL)
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					host = addr
				}
				if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
					return nil, fmt.Errorf("blocked private address %s", host)
				}
				d := net.Dialer{Timeout: timeout}
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}

// ValidateOutboundURL ensures URL targets an allowlisted public host.
func ValidateOutboundURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	return validateOutboundURL(u)
}

func validateOutboundURL(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("only https allowed")
	}
	host := strings.ToLower(u.Hostname())
	if _, ok := allowedHosts[host]; !ok {
		return fmt.Errorf("host %q not allowlisted", host)
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup %q: %w", host, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("blocked private ip for %q", host)
		}
	}
	return nil
}

// DoRequest performs an outbound request after SSRF validation.
func DoRequest(ctx context.Context, client *http.Client, method, rawURL string) (*http.Response, error) {
	if err := ValidateOutboundURL(rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
