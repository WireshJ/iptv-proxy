package server

import (
	"net/http"
	"testing"
)

func TestRepairAcceptRanges(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		status                int
		accept, content, want string
	}{
		{"provider numeric extent", 206, "0-311549952", "bytes 1048576-1052671/311549952", "bytes"},
		{"standard bytes", 206, "bytes", "bytes 0-4095/8192", "bytes"},
		{"no range support", 200, "none", "", "none"},
		{"no confirmed partial response", 200, "0-8192", "", "0-8192"},
		{"invalid content range", 206, "0-8192", "invalid", "0-8192"},
		{"missing advertisement", 206, "", "bytes 0-4095/8192", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			h.Set("Accept-Ranges", tc.accept)
			h.Set("Content-Range", tc.content)
			h.Set("Content-Length", "4096")
			repairAcceptRanges(tc.status, h)
			if h.Get("Accept-Ranges") != tc.want || h.Get("Content-Range") != tc.content || h.Get("Content-Length") != "4096" {
				t.Fatal("incorrect range metadata")
			}
		})
	}
}
