package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLayout Log layout
type LogLayout struct {
	Time      time.Time
	Metadata  map[string]interface{} // Store custom original data
	Path      string                 // Access path
	Query     string                 //Include query
	Body      string                 //Include body data
	IP        string                 // IP address
	UserAgent string                 //acting
	Error     string                 //mistake
	Cost      time.Duration          //Spend time
	Source    string                 //source
}

type Logger struct {
	// Filter User custom filter
	Filter func(c *gin.Context) bool
	// FilterKeyword Keyword filter(key)
	FilterKeyword func(layout *LogLayout) bool
	// AuthProcess Authentication processing
	AuthProcess func(c *gin.Context, layout *LogLayout)
	// Log processing
	Print func(LogLayout)
	// Source Service unique identifier
	Source string
}

func (l Logger) SetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		var body []byte
		if l.Filter != nil && !l.Filter(c) {
			body, _ = c.GetRawData()
			// Put the original body back
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
		cost := time.Since(start)
		layout := LogLayout{
			Time:      time.Now(),
			Path:      path,
			Query:     query,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Error:     strings.TrimRight(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"),
			Cost:      cost,
			Source:    l.Source,
		}
		if l.Filter != nil && !l.Filter(c) {
			layout.Body = string(body)
		}
		if l.AuthProcess != nil {
			// Process the information needed for authentication
			l.AuthProcess(c, &layout)
		}
		if l.FilterKeyword != nil {
			// Self-judge key/value desensitization etc.
			l.FilterKeyword(&layout)
		}
		// Self-process the log
		l.Print(layout)
	}
}

func DefaultLogger() gin.HandlerFunc {
	return Logger{
		Print: func(layout LogLayout) {
			// Standard output, k8s does collection
			v, _ := json.Marshal(layout)
			fmt.Println(string(v))
		},
		Source: "GVA",
	}.SetLoggerMiddleware()
}
