package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Some Xtream players request catch-up through the legacy query endpoint.
func (c *Config) xtreamStreamTimeshiftPHP(ctx *gin.Context) {
	q := ctx.Request.URL.Query()
	for _, key := range []string{"stream", "start", "duration"} {
		if q.Get(key) == "" {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	u, err := url.Parse(strings.TrimRight(c.XtreamBaseURL, "/") + "/streaming/timeshift.php")
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}
	q.Set("username", c.XtreamUser.String())
	q.Set("password", c.XtreamPassword.String())
	q.Set("start", normalizeTimeshiftStart(q.Get("start")))
	u.RawQuery = q.Encode()
	c.stream(ctx, u)
}
