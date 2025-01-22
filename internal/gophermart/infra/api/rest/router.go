package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gofermart/internal/gophermart/core/model"
)

type ServerService interface {
	UserRegister(ctx context.Context, request model.User) (string, error)
	UserLogin(ctx context.Context, request model.User) (string, error)
	ValidateToken(tokenString string) (string, error)

	UserOrder(ctx context.Context, userLogin, orderID string) error
	UserOrders(ctx context.Context, userLogin string) ([]model.Order, error)
}

type Config struct {
	Server ServerService
	Logger zap.SugaredLogger
	Port   int64
}

type Router struct {
	srv *http.Server
}

type handler struct {
	server ServerService
	logger zap.SugaredLogger
}

func NewRouter(conf Config) *Router {
	h := &handler{
		server: conf.Server,
		logger: conf.Logger,
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(h.mwDecompress())
	router.Use(h.responseGzipMiddleware())

	userGroup := router.Group("/api/user")
	{
		userGroup.POST("/register", h.userRegister)
		userGroup.POST("/login", h.userLogin)
	}

	ordersGroup := router.Group("/api/user/orders")
	{
		ordersGroup.Use(h.validationJWTMiddleware())
		ordersGroup.POST("", h.userOrder)
		ordersGroup.GET("", h.userOrders)
	}

	h.logger.Infof("server started on port: %d", conf.Port)

	return &Router{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", conf.Port),
			Handler: router,
		},
	}
}

func (r *Router) Run() error {
	if err := r.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}

	return nil
}
