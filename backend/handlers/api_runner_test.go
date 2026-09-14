package handlers

import (
	"context"
	"net/http"
	"net/netip"
	"testing"
)

type fakeResolver struct{ ips []netip.Addr }

func (f fakeResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return f.ips, nil
}
func TestRunnerRejectsPrivateAndUnallowlistedTargets(t *testing.T) {
	r := APIRunner{AllowedHosts: map[string]bool{"api.example.test": true}, Resolver: fakeResolver{ips: []netip.Addr{netip.MustParseAddr("127.0.0.1")}}}
	if _, err := r.Run(context.Background(), "GET", "http://api.example.test/x", "", nil); err == nil {
		t.Fatal("private DNS answer must be blocked")
	}
	if _, err := r.Run(context.Background(), "GET", "http://localhost/x", "", nil); err == nil {
		t.Fatal("localhost must be blocked before connection")
	}
}
func TestRunnerRedactsSensitiveHeaders(t *testing.T) {
	headers := redactHeaders(http.Header{"Set-Cookie": {"secret"}, "Authorization": {"Bearer secret"}, "Content-Type": {"application/json"}})
	if headers["Set-Cookie"] != "[REDACTED]" || headers["Authorization"] != "[REDACTED]" || headers["Content-Type"] != "application/json" {
		t.Fatalf("redaction failed: %#v", headers)
	}
}
