package handlers

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// DumpRequestBodyHandler gin handler - for debug purposes only
func DumpRequestBodyHandler(c *gin.Context) {
	b := bufio.NewReader(c.Request.Body)
	s, err := b.ReadString(0)
	slog.Info("default handler", "headers", c.Request.Header.Values("Authentication"), "body", s, "error", err)
	_ = c.AbortWithError(http.StatusNotImplemented, http.ErrAbortHandler)
}
