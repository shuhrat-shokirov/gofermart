package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gofermart/internal/gophermart/core/application"
	"gofermart/internal/gophermart/core/model"
)

func (h *handler) userWithdraw(c *gin.Context) {
	login := c.GetString(loginKey)

	var request model.WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("failed to bind request: %v", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.server.UserWithdraw(context.TODO(), login, request)
	if err != nil {
		if errors.Is(err, application.ErrInsufficientFunds) {
			c.Writer.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if errors.Is(err, application.ErrInvalidOrderID) {
			c.Writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		h.logger.Errorf("failed to withdraw: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func (h *handler) userWithdrawals(c *gin.Context) {
	login := c.GetString(loginKey)

	list, err := h.server.UserWithdrawals(context.TODO(), login)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.Writer.WriteHeader(http.StatusNoContent)
			return
		}

		h.logger.Errorf("failed to get user withdrawals: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, list)
}
