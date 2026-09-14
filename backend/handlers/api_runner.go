package handlers

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const maxRunnerBody = 256 * 1024

type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}
type APIRunner struct {
	AllowedHosts map[string]bool
	Resolver     Resolver
	Client       *http.Client
	MaxBody      int64
}
type RunnerResponse struct {
	StatusCode int
	Duration   time.Duration
	Size       int
	Headers    map[string]string
	Body       string
	Truncated  bool
}

func privateAddress(ip net.IP) bool {
	a, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	a = a.Unmap()
	return a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() || a.IsMulticast() || a.IsUnspecified()
}

func (r APIRunner) validateURL(ctx context.Context, raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, fmt.Errorf("runner only permits absolute http/https URLs")
	}
	host := strings.ToLower(u.Hostname())
	if !r.AllowedHosts[host] {
		return nil, fmt.Errorf("target host is not allowlisted")
	}
	if ip := net.ParseIP(host); ip != nil {
		if privateAddress(ip) {
			return nil, fmt.Errorf("private target is blocked")
		}
		return u, nil
	}
	resolver := r.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ips, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("target DNS resolution failed")
	}
	for _, ip := range ips {
		if privateAddress(net.IP(ip.AsSlice())) {
			return nil, fmt.Errorf("private target is blocked")
		}
	}
	return u, nil
}

func redactHeaders(headers http.Header) map[string]string {
	out := map[string]string{}
	for key, values := range headers {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "cookie" || lower == "set-cookie" || strings.Contains(lower, "api-key") {
			out[key] = "[REDACTED]"
		} else {
			out[key] = strings.Join(values, ", ")
		}
	}
	return out
}

func (r APIRunner) Run(ctx context.Context, method, rawURL, body string, headers map[string]string) (RunnerResponse, error) {
	u, err := r.validateURL(ctx, rawURL)
	if err != nil {
		return RunnerResponse{}, err
	}
	if len(body) > maxRunnerBody {
		return RunnerResponse{}, fmt.Errorf("request body exceeds limit")
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(body))
	if err != nil {
		return RunnerResponse{}, err
	}
	for key, value := range headers {
		if strings.EqualFold(key, "host") || strings.EqualFold(key, "cookie") {
			continue
		}
		req.Header.Set(key, value)
	}
	client := r.Client
	if client == nil {
		client = &http.Client{}
	}
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	start := time.Now()
	response, err := clone.Do(req)
	if err != nil {
		return RunnerResponse{}, err
	}
	defer response.Body.Close()
	limit := r.MaxBody
	if limit <= 0 {
		limit = maxRunnerBody
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return RunnerResponse{}, err
	}
	truncated := int64(len(data)) > limit
	if truncated {
		data = data[:limit]
	}
	return RunnerResponse{StatusCode: response.StatusCode, Duration: time.Since(start), Size: len(data), Headers: redactHeaders(response.Header), Body: string(data), Truncated: truncated}, nil
}
