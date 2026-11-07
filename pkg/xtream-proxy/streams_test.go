package xtreamproxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	xtream "github.com/tellytv/go.xtream-codes"
)

func TestStreamCategoryIDAcceptsNumberAndString(t *testing.T) {
	for _, action := range []string{"get_vod_streams", "get_live_streams"} {
		for _, tc := range []struct {
			name      string
			body      string
			wantError bool
		}{
			{"numeric", `[{"stream_id":1,"category_id":123}]`, false},
			{"quoted", `[{"stream_id":1,"category_id":"123"}]`, false},
			{"mixed", `[{"stream_id":1,"category_id":123},{"stream_id":2,"category_id":"123"}]`, false},
			{"invalid", `[{"stream_id":1,"category_id":"not-an-id"}]`, true},
		} {
			t.Run(action+"/"+tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Query().Get("action") == "" {
						_, _ = w.Write([]byte(`{"user_info":{"auth":1},"server_info":{}}`))
						return
					}
					if r.URL.Query().Get("action") != action || r.URL.Query().Get("category_id") != "123" {
						t.Error("stream action or category filter changed")
					}
					_, _ = w.Write([]byte(tc.body))
				}))
				defer server.Close()
				client, err := xtream.NewClientWithUserAgent(context.Background(), "fixture-user", "fixture-password", server.URL, "fixture-agent")
				if err != nil {
					t.Fatal(err)
				}
				var streams []xtream.Stream
				if action == "get_vod_streams" {
					streams, err = client.GetVideoOnDemandStreams("123")
				} else {
					streams, err = client.GetLiveStreams("123")
				}
				if tc.wantError {
					if err == nil {
						t.Fatal("invalid category accepted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				var original []json.RawMessage
				if err := json.Unmarshal([]byte(tc.body), &original); err != nil {
					t.Fatal(err)
				}
				if len(streams) != len(original) {
					t.Fatalf("lost streams: got %d, want %d", len(streams), len(original))
				}
				for _, stream := range streams {
					if stream.CategoryID != 123 {
						t.Fatalf("category changed: %d", stream.CategoryID)
					}
				}
			})
		}
	}
}
