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

type epgFixtureTransport func(*http.Request) (*http.Response, error)

func (f epgFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestXtreamEPGHTTPResponse(t *testing.T) {
	const listing = `{"epg_listings":[{"title":"VGVzdA==","description":"w7g/","has_archive":1,"stream_id":"540869"}]}`
	for _, tc := range []struct {
		name, providerBody string
		wantStatus         int
		wantBody           string
	}{
		{"archive", listing, http.StatusOK, listing},
		{"provider-invalid-json", "", http.StatusBadGateway, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previousClient := http.DefaultClient
			t.Cleanup(func() { http.DefaultClient = previousClient })
			http.DefaultClient = &http.Client{Transport: epgFixtureTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "provider.invalid" {
					t.Fatal("unexpected destination")
				}
				body := tc.providerBody
				if r.URL.Query().Get("action") == "" {
					body = `{"user_info":{"auth":1},"server_info":{}}`
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			server := &Config{ProxyConfig: &config.ProxyConfig{
				XtreamBaseURL: "http://provider.invalid", XtreamUser: "fixture-user", XtreamPassword: "fixture-password",
			}}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "http://proxy.invalid/player_api.php?action=get_simple_data_table&stream_id=540869", nil)
			server.xtreamPlayerAPIGET(ctx)
			if recorder.Code != tc.wantStatus || recorder.Body.String() != tc.wantBody {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
			if tc.wantStatus == http.StatusOK && !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
				t.Fatal("missing JSON content type")
			}
		})
	}
}
