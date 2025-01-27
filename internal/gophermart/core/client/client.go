package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gofermart/internal/gophermart/core/model"

	"github.com/imroc/req/v3"
)

type Client interface {
	SendOrder(ctx context.Context, orderID string) (model.ClientResponse, error)
}

type handler struct {
	client     *req.Client
	ch         chan struct{}
	retryAfter *time.Duration
}

func NewClient(addr string, limit int64) Client {
	const timeout = 30 * time.Second

	return &handler{
		client: req.NewClient().
			SetBaseURL(fmt.Sprintf("http://%s/api/orders", addr)).
			SetCommonContentType("application/json").
			SetTimeout(timeout),
		ch: make(chan struct{}, limit),
	}
}

func (c *handler) SendOrder(ctx context.Context, orderID string) (model.ClientResponse, error) {
	if c.retryAfter != nil {
		time.Sleep(*c.retryAfter)
		c.retryAfter = nil
	}

	c.ch <- struct{}{}
	defer func() {
		<-c.ch
	}()

	response, err := c.client.
		R().
		SetContext(ctx).
		Get("/" + orderID)
	if err != nil {
		return model.ClientResponse{}, fmt.Errorf("can't send order: %w", err)
	}

	if response.Response.StatusCode != http.StatusOK {
		if response.Response.StatusCode == http.StatusNoContent {
			return model.ClientResponse{}, fmt.Errorf("order not registered, %w", ErrNotFound)
		}

		if response.Response.StatusCode == http.StatusTooManyRequests {
			retryAfter := response.Response.Header.Get("Retry-After")
			if retryAfter != "" {
				duration, err := time.ParseDuration(retryAfter + "s")
				if err != nil {
					return model.ClientResponse{}, fmt.Errorf("can't parse Retry-After header: %w", err)
				}

				c.retryAfter = &duration
			}

			return model.ClientResponse{}, fmt.Errorf("too many requests, %w", ErrTooManyRequests)
		}

		return model.ClientResponse{}, fmt.Errorf("can't send order: status code %d", response.Response.StatusCode)
	}

	var clientResponse model.ClientResponse
	err = response.UnmarshalJson(&clientResponse)
	if err != nil {
		return model.ClientResponse{}, fmt.Errorf("can't unmarshal response: %w", err)
	}

	return clientResponse, nil
}
