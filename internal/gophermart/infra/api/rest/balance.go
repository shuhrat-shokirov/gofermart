package rest

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

const loginKey = "login"

func (h *handler) userBalance(c *gin.Context) {
	login := c.GetString(loginKey)

	balance, err := h.server.UserBalance(context.TODO(), login)
	if err != nil {
		h.logger.Errorf("failed to get balance: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Infof("User %s has balance: %v", login, balance)

	c.JSON(http.StatusOK, balance)
}
