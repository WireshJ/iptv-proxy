package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WireshJ/iptv-proxy/pkg/config"
	"github.com/gin-gonic/gin"
)

func TestNormalizeTimeshiftStart(t *testing.T) {
	for input, want := range map[string]string{
		"2026-10-10:08-55 AM":    "2026-10-10:08-55",
		"2026-10-10:08-55 PM":    "2026-10-10:20-55",
		"2026-10-10:12-00 AM":    "2026-10-10:00-00",
		"2026-10-10:12-00 PM":    "2026-10-10:12-00",
		"2026-10-10:08:55 pm":    "2026-10-10:20-55",
		"2026-10-10:08:55:00 AM": "2026-10-10:08:55:00",
		"2026-10-10:8-55PM":      "2026-10-10:20-55",
		"2026-10-10 20:55":       "2026-10-10:20-55",
		"2026-10-10T20:55:31":    "2026-10-10:20:55:31",
		"2026-10-10:00-00":       "2026-10-10:00-00",
		"2026-10-10:12-00":       "2026-10-10:12-00",
		"2026-10-10:23-59":       "2026-10-10:23-59",
		"2026-10-10:08-55":       "2026-10-10:08-55",
		"1791615300":             "1791615300",
		"2026-10-10:18-55 PM":    "2026-10-10:18-55 PM",
	} {
		if got := normalizeTimeshiftStart(input); got != want {
			t.Errorf("normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTimeshiftPathWithAMClock(t *testing.T) {
	previous := streamClient
	t.Cleanup(func() { streamClient = previous })
	streamClient = &http.Client{Transport: epgFixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "provider.invalid" || r.URL.Path != "/timeshift/provider-user/provider-password/10/2026-10-10:08-55/540869.ts" {
			t.Fatal("incorrect provider timeshift request")
		}
		if r.Header.Get("Range") != "bytes=0-" {
			t.Fatal("Range header lost")
		}
		return &http.Response{StatusCode: http.StatusPartialContent, Header: http.Header{"Content-Type": {"video/mp2t"}}, Body: io.NopCloser(strings.NewReader("fixture-video"))}, nil
	})}
	c := &Config{ProxyConfig: &config.ProxyConfig{XtreamBaseURL: "http://provider.invalid", XtreamUser: "provider-user", XtreamPassword: "provider-password", User: "proxy-user", Password: "proxy-password"}}
	router := gin.New()
	c.xtreamRoutes(router.Group("/"))
	req := httptest.NewRequest(http.MethodGet, "http://proxy.invalid/timeshift/proxy-user/proxy-password/10/2026-10-10:08-55%20AM/540869.ts", nil)
	req.Header.Set("Range", "bytes=0-")
	recorder := newStreamRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusPartialContent || recorder.Body.String() != "fixture-video" {
		t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}
