package krakend

import (
	"io"
	"net/http"
	"os"
	"sync"

	gin_logger "github.com/Unacademy/krakend-gin-logger"

	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	httpsecure "github.com/devopsfaith/krakend-httpsecure/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
)

var (
	httpClient *http.Client
	clientOnce sync.Once
)

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin_logger.NewLogger(cfg.ExtraConfig, logger, gin.LoggerConfig{Output: w}), gin.Recovery())

	engine.NoRoute(handleNoMatch)

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	if err := httpsecure.Register(cfg.ExtraConfig, engine); err != nil {
		logger.Warning(err)
	}

	lua.Register(logger, cfg.ExtraConfig, engine)

	botdetector.Register(cfg, logger, engine)

	return engine
}

type engineFactory struct{}

func (e engineFactory) NewEngine(cfg config.ServiceConfig, l logging.Logger, w io.Writer) *gin.Engine {
	return NewEngine(cfg, l, w)
}

func getClient() *http.Client {
	clientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        100, // default 100
			MaxIdleConnsPerHost: 100, // default 2
		}

		httpClient = &http.Client{
			Transport: transport,
		}
	})

	return httpClient
}

func handleNoMatch(c *gin.Context) {
	client := getClient()
	req, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header = c.Request.Header

	// req.URL.Scheme = viper.GetString("DEFAULT_URL_SCHEME") // should be either http or https for current use case
	// req.URL.Host = viper.GetString("DEFAULT_URL_HOST")

	req.URL.Scheme = os.Getenv("DEFAULT_URL_SCHEME") // should be either http or https for current use case
	req.URL.Host = os.Getenv("DEFAULT_URL_HOST")

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
