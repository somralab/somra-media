package outbound

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PinnedClient performs HTTP requests only to the host configured at construction.
type PinnedClient struct {
	base              *url.URL
	httpClient        *http.Client
	allowPrivateHosts bool
}

// ClientOption configures PinnedClient construction.
type ClientOption func(*PinnedClient)

// AllowPrivateHosts disables private-IP blocking (tests only).
func AllowPrivateHosts() ClientOption {
	return func(c *PinnedClient) { c.allowPrivateHosts = true }
}

// NewPinnedClient binds outbound requests to baseURL's scheme and host.
func NewPinnedClient(baseURL string, timeout time.Duration, opts ...ClientOption) (*PinnedClient, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("pinned client base url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("pinned client base url: scheme and host required")
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	pc := &PinnedClient{
		base: u,
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return validateSameHost(u, req.URL)
			},
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					d := net.Dialer{Timeout: timeout}
					return d.DialContext(ctx, network, addr)
				},
			},
		},
	}
	for _, opt := range opts {
		opt(pc)
	}
	return pc, nil
}

// Get performs a GET relative to the pinned base URL.
func (c *PinnedClient) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, query, nil)
}

// PostForm performs POST with application/x-www-form-urlencoded body.
func (c *PinnedClient) PostForm(ctx context.Context, path string, form url.Values) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, nil, strings.NewReader(form.Encode()))
}

func (c *PinnedClient) do(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	if c == nil || c.base == nil {
		return nil, fmt.Errorf("pinned client: not initialized")
	}
	target, err := c.resolve(path, query)
	if err != nil {
		return nil, err
	}
	if err := validateSameHost(c.base, target); err != nil {
		return nil, err
	}
	if !c.allowPrivateHosts {
		if err := blockPrivateHost(target.Hostname()); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, err
	}
	if method == http.MethodPost && body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return c.httpClient.Do(req)
}

func (c *PinnedClient) resolve(path string, query url.Values) (*url.URL, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("pinned client path: %w", err)
	}
	target := c.base.ResolveReference(ref)
	if query != nil {
		q := target.Query()
		for k, vals := range query {
			for _, v := range vals {
				q.Add(k, v)
			}
		}
		target.RawQuery = q.Encode()
	}
	return target, nil
}

func validateSameHost(base, target *url.URL) error {
	if base.Scheme != target.Scheme {
		return fmt.Errorf("redirect scheme mismatch")
	}
	if !strings.EqualFold(base.Hostname(), target.Hostname()) {
		return fmt.Errorf("host mismatch")
	}
	return nil
}

func blockPrivateHost(host string) error {
	if host == "" {
		return fmt.Errorf("empty host")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("blocked private address %s", host)
		}
		return nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup %q: %w", host, err)
	}
	for _, addr := range addrs {
		if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() {
			return fmt.Errorf("blocked private ip for %q", host)
		}
	}
	return nil
}
