package krakend

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	metrics "github.com/devopsfaith/krakend-metrics/gin"
	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
)

// SSEConfig holds the configuration for SSE endpoints
type SSEConfig struct {
	KeepAliveInterval time.Duration `json:"keep_alive_interval"`
	RetryInterval     int           `json:"retry_interval"`
}

// SSEHandlerFactory creates handlers for SSE endpoints
type SSEHandlerFactory struct {
	logger  logging.Logger
	metrics *metrics.Metrics
}

// NewSSEHandlerFactory returns a new SSEHandlerFactory
func NewSSEHandlerFactory(logger logging.Logger, metrics *metrics.Metrics) *SSEHandlerFactory {
	return &SSEHandlerFactory{
		logger:  logger,
		metrics: metrics,
	}
}

// NewHandler creates a new SSE handler
// NewHandler creates a new SSE handler
func (s *SSEHandlerFactory) NewHandler(cfg *config.EndpointConfig, prxy proxy.Proxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// Get SSE config
		var sseCfg SSEConfig
		if v, ok := cfg.ExtraConfig["sse"]; ok && v != nil {
			if b, err := json.Marshal(v); err == nil {
				json.Unmarshal(b, &sseCfg)
			}
		}

		// Set default values
		if sseCfg.KeepAliveInterval == 0 {
			sseCfg.KeepAliveInterval = 30 * time.Second
		}
		if sseCfg.RetryInterval == 0 {
			sseCfg.RetryInterval = 1000
		}

		// Send retry interval
		c.Writer.Write([]byte("retry: " + strconv.Itoa(sseCfg.RetryInterval) + "\n\n"))
		c.Writer.Flush()

		// Start keep-alive goroutine in a separate context
		keepAliveCtx, keepAliveCancel := context.WithCancel(context.Background())
		defer keepAliveCancel()

		go func() {
			ticker := time.NewTicker(sseCfg.KeepAliveInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					c.Writer.Write([]byte(": keepalive\n\n"))
					c.Writer.Flush()
				case <-keepAliveCtx.Done():
					return
				}
			}
		}()

		// Manually construct the backend URL
		if len(cfg.Backend) == 0 {
			s.logger.Error("No backend configured for SSE endpoint")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		backendConfig := cfg.Backend[0]
		if len(backendConfig.Host) == 0 {
			s.logger.Error("No host configured for SSE backend")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Construct the backend URL
		backendURL := fmt.Sprintf("%s%s",
			backendConfig.Host[0],
			backendConfig.URLPattern)

		s.logger.Debug(fmt.Sprintf("SSE backend URL: %s", backendURL))

		// Create a new HTTP client
		client := &http.Client{
			Timeout: 0, // No timeout for streaming connections
		}

		// Create request
		req, err := http.NewRequestWithContext(c.Request.Context(),
			backendConfig.Method,
			backendURL,
			c.Request.Body)

		if err != nil {
			s.logger.Error("Error creating backend request:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Copy relevant headers
		for k, v := range c.Request.Header {
			req.Header[k] = v
		}

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			s.logger.Error("Error making backend request:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			s.logger.Error(fmt.Sprintf("Backend returned non-200 status: %d", resp.StatusCode))
			c.AbortWithStatus(resp.StatusCode)
			return
		}

		// Stream the response
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					s.logger.Error("SSE read error:", err)
				}
				break
			}

			// Write the line directly to the client
			c.Writer.Write(line)
			c.Writer.Flush()
		}
	}
}
