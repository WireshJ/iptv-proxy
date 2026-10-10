package xtreamproxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/WireshJ/iptv-proxy/pkg/config"
	xtream "github.com/tellytv/go.xtream-codes"
)

type epgTransport func(*http.Request) (*http.Response, error)

func (f epgTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func epgTestClient(t *testing.T, status int, body string, check func(*http.Request)) *Client {
	t.Helper()
	return &Client{&xtream.XtreamClient{
		Username: "fixture-user", Password: "fixture-password", BaseURL: "http://provider.invalid",
		UserAgent: "fixture-agent", Context: context.Background(),
		HTTP: &http.Client{Transport: epgTransport(func(r *http.Request) (*http.Response, error) {
			if check != nil {
				check(r)
			}
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})},
	}}
}

func TestEPGPreservesProviderEnvelopeAndEncodings(t *testing.T) {
	const body = `{"epg_listings":[{"id":"17","title":"VGVzdA==","description":"w7g/","has_archive":"1","start_timestamp":"1791612000","provider_extra":{"keep":true}}],"provider_metadata":true}`
	for _, action := range []string{getSimpleDataTable, getShortEPG} {
		t.Run(action, func(t *testing.T) {
			c := epgTestClient(t, http.StatusOK, body, func(r *http.Request) {
				q := r.URL.Query()
				if r.URL.Host != "provider.invalid" || r.URL.Path != "/player_api.php" || q.Get("action") != action || q.Get("stream_id") != "540869" || q.Get("username") != "fixture-user" || q.Get("password") != "fixture-password" || r.Header.Get("User-Agent") != "fixture-agent" {
					t.Fatalf("incorrect EPG request")
				}
				if action == getShortEPG && q.Get("limit") != "8" {
					t.Fatal("short EPG limit lost")
				}
			})
			q := url.Values{"stream_id": {"540869"}}
			if action == getShortEPG {
				q.Set("limit", "8")
			}
			response, _, err := c.Action(&config.ProxyConfig{}, action, q)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != body {
				t.Fatalf("EPG response changed: %s", encoded)
			}
		})
	}
}

func TestEPGErrorsReturnBadGateway(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"empty", 200, ""}, {"html", 200, "<html>failure</html>"},
		{"missing-listings", 200, `{}`}, {"wrong-listings-type", 200, `{"epg_listings":{}}`},
		{"provider-failure", 503, `{"epg_listings":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := epgTestClient(t, tc.status, tc.body, nil)
			_, status, err := c.Action(&config.ProxyConfig{}, getSimpleDataTable, url.Values{"stream_id": {"540869"}})
			if err == nil || status != http.StatusBadGateway {
				t.Fatalf("status=%d error=%v", status, err)
			}
			if strings.Contains(err.Error(), "fixture-password") {
				t.Fatal("credentials in error")
			}
		})
	}
}

func TestEPGEmptyArchiveRemainsAnEnvelope(t *testing.T) {
	c := epgTestClient(t, 200, `[]`, nil)
	response, _, err := c.Action(&config.ProxyConfig{}, getSimpleDataTable, url.Values{"stream_id": {"540869"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil || string(encoded) != `{"epg_listings":[]}` {
		t.Fatalf("empty envelope: %s, %v", encoded, err)
	}
}

func TestEPGEnvelopeWithoutBase64Padding(t *testing.T) {
	const body = `{"epg_listings":[{"title":"VGVzdA","description":"dGVzdA","has_archive":"1"}],"provider_metadata":true}`
	c := epgTestClient(t, 200, body, nil)
	response, _, err := c.Action(&config.ProxyConfig{}, getSimpleDataTable, url.Values{"stream_id": {"540869"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil || string(encoded) != body {
		t.Fatalf("provider envelope changed: %s, %v", encoded, err)
	}
}

func TestEPGMissingStreamDoesNotContactProvider(t *testing.T) {
	c := epgTestClient(t, 200, `{"epg_listings":[]}`, func(*http.Request) { t.Fatal("unexpected provider request") })
	_, status, err := c.Action(&config.ProxyConfig{}, getSimpleDataTable, url.Values{})
	if err == nil || status != http.StatusBadRequest {
		t.Fatalf("status=%d error=%v", status, err)
	}
}

func TestLoginAdvertisesHostnameWithoutScheme(t *testing.T) {
	c := epgTestClient(t, 200, `{}`, func(*http.Request) { t.Fatal("unexpected provider request") })
	response, _, err := c.Action(&config.ProxyConfig{HostConfig: &config.HostConfiguration{Hostname: "proxy.invalid"}, AdvertisedPort: 9191}, "", url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		ServerInfo struct {
			URL      string `json:"url"`
			Protocol string `json:"server_protocol"`
		} `json:"server_info"`
	}
	if err := json.Unmarshal(encoded, &data); err != nil {
		t.Fatal(err)
	}
	if data.ServerInfo.URL != "proxy.invalid" || data.ServerInfo.Protocol != "http" {
		t.Fatalf("invalid advertised server: %s", encoded)
	}
}
