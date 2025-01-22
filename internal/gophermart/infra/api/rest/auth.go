package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gofermart/internal/gophermart/core/application"
	"gofermart/internal/gophermart/core/model"
)

const (
	cookieMaxAge = 3600
)

func (h *handler) userRegister(c *gin.Context) {
	var request model.User
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("failed to bind request: %v", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.server.UserRegister(context.TODO(), request)
	if err != nil {
		if errors.Is(err, application.ErrUserExists) {
			c.Writer.WriteHeader(http.StatusConflict)
			return
		}

		h.logger.Errorf("failed to register user: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}

	c.SetCookie("Authorization", token, cookieMaxAge, "/", "", false, true)

	c.Writer.WriteHeader(http.StatusOK)
}

func (h *handler) userLogin(c *gin.Context) {
	var request model.User
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("failed to bind request: %v", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.server.UserLogin(context.TODO(), request)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) ||
			errors.Is(err, application.ErrIncorrectPass) {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.logger.Errorf("failed to login user: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}

	c.SetCookie("Authorization", token, cookieMaxAge, "/", "", false, true)

	c.Writer.WriteHeader(http.StatusOK)
}
