package rest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	Logger zap.SugaredLogger
	Port   int64
}

type Router struct {
	srv *http.Server
}

type handler struct {
	logger  zap.SugaredLogger
	hashKey string
}

func NewRouter(conf Config) *Router {
	h := &handler{
		logger: conf.Logger,
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(h.mwDecompress())
	router.Use(h.responseGzipMiddleware())
	router.Use(h.encryptionMiddleware())

	h.logger.Infof("server started on port: %d", conf.Port)

	return &Router{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", conf.Port),
			Handler: router,
		},
	}
}

func (a *Router) Run() error {
	if err := a.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}

	return nil
}
