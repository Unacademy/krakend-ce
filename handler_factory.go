package krakend

import (
	botdetector "github.com/devopsfaith/krakend-botdetector/gin"
	jose "github.com/devopsfaith/krakend-jose"
	ginjose "github.com/devopsfaith/krakend-jose/gin"
	lua "github.com/devopsfaith/krakend-lua/router/gin"
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	opencensus "github.com/devopsfaith/krakend-opencensus/router/gin"
	juju "github.com/devopsfaith/krakend-ratelimit/juju/router/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	router "github.com/luraproject/lura/router/gin"
	krakendauth "github.com/unacademy/krakend-auth"
	sse "github.com/unacademy/krakend-sse"
	websocket "github.com/unacademy/krakend-websocket"
)

type handlerFactory struct {
	serviceConfig config.ServiceConfig
}

func (h handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactoryWithConfig(l, m, r, h.serviceConfig)
}

// NewHandlerFactoryWithConfig returns a HandlerFactory with service configuration for WebSocket backends
func NewHandlerFactoryWithConfig(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory, serviceConfig config.ServiceConfig) router.HandlerFactory {
	handlerFactory := juju.HandlerFactory
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = krakendauth.HandlerFactory(handlerFactory, logger)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)

	// Wrap with SSE middleware
	handlerFactory = sse.New(handlerFactory, logger)

	// Wrap with WebSocket middleware - this should be the outermost wrapper
	// so it can intercept WebSocket upgrade requests before any other middleware
	handlerFactory = websocket.New(handlerFactory, logger)

	return handlerFactory
}
