package handleNoMatch

import (
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	client     *http.Client
	clientOnce sync.Once
)

func getClient() *http.Client {
	clientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        100, // default 100
			MaxIdleConnsPerHost: 100, // default 2
		}

		client = &http.Client{
			Transport: transport,
		}
	})

	return client
}

func handleNoMatch(c *gin.Context) {
	client = getClient()
	req, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header = c.Request.Header

	req.URL.Scheme = viper.GetString("DEFAULT_URL_SCHEME") // should be either http or https for current use case
	req.URL.Host = viper.GetString("DEFAULT_URL_HOST")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request"})
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
