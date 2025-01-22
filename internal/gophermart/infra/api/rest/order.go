package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gofermart/internal/gophermart/core/application"
)

func (h *handler) userOrder(c *gin.Context) {
	if c.ContentType() != "text/plain" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		h.logger.Errorf("failed to read body: %v", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	login := c.GetString("login")

	err = h.server.UserOrder(context.TODO(), login, string(body))
	if err != nil {
		if errors.Is(err, application.ErrOrderAlreadyExists) {
			c.Writer.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, application.ErrOrderExistsOnAnotherUser) {
			c.Writer.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, application.ErrInvalidOrderID) {
			c.Writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		h.logger.Errorf("failed to create order: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusAccepted)
}

func (h *handler) userOrders(c *gin.Context) {
	login := c.GetString("login")

	orders, err := h.server.UserOrders(context.TODO(), login)
	if err != nil {
		h.logger.Errorf("failed to get orders: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
