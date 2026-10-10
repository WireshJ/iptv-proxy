package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/WireshJ/iptv-proxy/pkg/config"
	"github.com/gin-gonic/gin"
)

type streamRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (r *streamRecorder) CloseNotify() <-chan bool { return r.closed }

func newStreamRecorder() *streamRecorder {
	return &streamRecorder{httptest.NewRecorder(), make(chan bool)}
}

func TestTimeshiftQueryEndpoint(t *testing.T) {
	for _, endpoint := range []string{"/streaming/timeshift.php", "/timeshift.php"} {
		t.Run(endpoint, func(t *testing.T) {
			previous := streamClient
			t.Cleanup(func() { streamClient = previous })
			streamClient = &http.Client{Transport: epgFixtureTransport(func(r *http.Request) (*http.Response, error) {
				q := r.URL.Query()
				if r.URL.Host != "provider.invalid" || r.URL.Path != "/streaming/timeshift.php" || q.Get("username") != "provider-user" || q.Get("password") != "provider-password" || q.Get("stream") != "540869" || q.Get("duration") != "50" || q.Get("start") != "2026-10-10:08-55" {
					t.Fatal("incorrect provider request")
				}
				if r.Header.Get("Range") != "bytes=0-4095" {
					t.Fatal("Range header lost")
				}
				return &http.Response{StatusCode: http.StatusPartialContent, Header: http.Header{"Content-Type": {"video/mp2t"}}, Body: io.NopCloser(strings.NewReader("fixture-video"))}, nil
			})}
			c := &Config{ProxyConfig: &config.ProxyConfig{XtreamBaseURL: "http://provider.invalid", XtreamUser: "provider-user", XtreamPassword: "provider-password", User: "proxy-user", Password: "proxy-password"}}
			router := gin.New()
			c.xtreamRoutes(router.Group("/"))
			q := url.Values{"username": {"proxy-user"}, "password": {"proxy-password"}, "stream": {"540869"}, "start": {"2026-10-10:08-55"}, "duration": {"50"}}
			req := httptest.NewRequest(http.MethodGet, "http://proxy.invalid"+endpoint+"?"+q.Encode(), nil)
			req.Header.Set("Range", "bytes=0-4095")
			recorder := newStreamRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusPartialContent || recorder.Body.String() != "fixture-video" {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
			q.Set("start", "2026-10-10:08-55 AM")
			req = httptest.NewRequest(http.MethodGet, "http://proxy.invalid"+endpoint+"?"+q.Encode(), nil)
			req.Header.Set("Range", "bytes=0-4095")
			recorder = newStreamRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusPartialContent {
				t.Fatalf("AM clock status=%d", recorder.Code)
			}
			q.Set("password", "incorrect")
			recorder = newStreamRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://proxy.invalid"+endpoint+"?"+q.Encode(), nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("unauthenticated status=%d", recorder.Code)
			}
			q.Set("password", "proxy-password")
			q.Del("start")
			recorder = newStreamRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://proxy.invalid"+endpoint+"?"+q.Encode(), nil))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("invalid query status=%d", recorder.Code)
			}
		})
	}
}
