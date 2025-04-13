package krakend

import (
	"fmt"

	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	jose "github.com/devopsfaith/krakend-jose"
	ginjose "github.com/devopsfaith/krakend-jose/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	opencensus "github.com/devopsfaith/krakend-opencensus/router/gin"
	juju "github.com/devopsfaith/krakend-ratelimit/juju/router/gin"
	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	router "github.com/luraproject/lura/router/gin"
	krakendauth "github.com/unacademy/krakend-auth"
)

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := juju.HandlerFactory
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = krakendauth.HandlerFactory(handlerFactory, logger)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)

	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the http handler", cfg.Endpoint))

		// Check if this is an SSE endpoint
		if _, ok := cfg.ExtraConfig["sse"]; ok {
			sseFactory := NewSSEHandlerFactory(logger, metricCollector)
			return sseFactory.NewHandler(cfg, p)
		}

		return handlerFactory(cfg, p)
	}
}

type handlerFactory struct{}

func (h handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactory(l, m, r)
}
