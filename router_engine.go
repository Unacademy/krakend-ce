package krakend

import (
	"io"
	"os"

	gin_logger "github.com/Unacademy/krakend-gin-logger"

	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	"github.com/devopsfaith/krakend-ce/pkg/customNoRouteHandler"
	httpsecure "github.com/devopsfaith/krakend-httpsecure/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
)

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin_logger.NewLogger(cfg.ExtraConfig, logger, gin.LoggerConfig{Output: w}), gin.Recovery())

	hanldeNoMatch := customNoRouteHandler.NewNoRouteHandler(os.Getenv("DEFAULT_URL_HOST"), os.Getenv("DEFAULT_URL_SCHEME"), logger)
	engine.NoRoute(hanldeNoMatch.ForwardRequestToDefaultURL)

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
