package xtreamproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Keep the provider's envelope and field encodings. The typed Xtream decoder
// rejects padded Base64 and returns only the listings slice, breaking clients.
func (c *Client) epgResponse(action, streamID string, limit int) (json.RawMessage, error) {
	u, err := url.Parse(strings.TrimRight(c.BaseURL, "/") + "/player_api.php")
	if err != nil {
		return nil, errors.New("invalid EPG provider address")
	}
	q := u.Query()
	q.Set("username", c.Username)
	q.Set("password", c.Password)
	q.Set("action", action)
	q.Set("stream_id", streamID)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(c.Context, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.New("could not build EPG request")
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, errors.New("EPG provider request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("EPG provider returned an unsuccessful response")
	}
	const maxEPGBytes = 16 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxEPGBytes+1))
	if err != nil || len(body) > maxEPGBytes {
		return nil, errors.New("could not read EPG provider response")
	}
	body = bytes.TrimSpace(body)
	if !json.Valid(body) {
		return nil, errors.New("EPG provider returned invalid JSON")
	}
	if bytes.Equal(body, []byte("[]")) {
		return json.RawMessage(`{"epg_listings":[]}`), nil
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, errors.New("EPG provider returned an invalid envelope")
	}
	listings := bytes.TrimSpace(envelope["epg_listings"])
	if len(listings) == 0 || (listings[0] != '[' && !bytes.Equal(listings, []byte("null"))) {
		return nil, errors.New("EPG provider response has no listings array")
	}
	return json.RawMessage(body), nil
}
