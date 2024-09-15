package handlers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// DumpResponseBodyHandler gin handler - for debug purposes only
func DumpResponseBodyHandler(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	statusCode := c.Writer.Status()
	slog.Info("Response logging", "status code", statusCode, "body", blw.body.String())
}
