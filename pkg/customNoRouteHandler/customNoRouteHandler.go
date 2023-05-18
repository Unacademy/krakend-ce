package customNoRouteHandler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/logging"
)

type NoRouteHandler struct {
	httpClient    *http.Client
	defaultURL    string
	defaultScheme string
	logger        logging.Logger
}

func NewNoRouteHandler(defaultURL, defaultScheme string, logger logging.Logger) *NoRouteHandler {
	h := &NoRouteHandler{
		defaultURL:    defaultURL,
		defaultScheme: defaultScheme,
		logger:        logger,
	}

	h.InitialiseHttpClient()

	return h
}

func (h *NoRouteHandler) InitialiseHttpClient() {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	h.httpClient = &http.Client{
		Transport: transport,
	}
}

func (h *NoRouteHandler) GetClient() *http.Client {
	return h.httpClient
}

func (h *NoRouteHandler) ForwardRequestToDefaultURL(c *gin.Context) {
	client := h.GetClient()
	req, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to create request:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header = c.Request.Header
	req.URL.Scheme = h.defaultScheme
	req.URL.Host = h.defaultURL

	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to forward request:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request"})
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	c.Status(resp.StatusCode)
	_, copyErr := io.Copy(c.Writer, resp.Body)
	if err != nil {
		h.logger.Error("Failed to copy response:", copyErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy response"})
		return
	}
}
