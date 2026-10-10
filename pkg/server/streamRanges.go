package server

import (
	"net/http"
	"regexp"
)

var numericAcceptRanges = regexp.MustCompile(`^[0-9]+-[0-9]+$`)
var byteContentRange = regexp.MustCompile(`^bytes [0-9]+-[0-9]+/(?:[0-9]+|\*)$`)

// Some providers put a numeric extent in Accept-Ranges instead of the range
// unit. Repair that known malformed value only when a byte response confirms
// that the upstream supports ranges (RFC 9110, section 14.3).
func repairAcceptRanges(status int, header http.Header) {
	if status == http.StatusPartialContent && numericAcceptRanges.MatchString(header.Get("Accept-Ranges")) && byteContentRange.MatchString(header.Get("Content-Range")) {
		header.Set("Accept-Ranges", "bytes")
	}
}
