package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"resty.dev/v3"
)

// HTTPAccrualClient http implementation AccrualClient
type HTTPAccrualClient struct {
	client *resty.Client
}

// NewHTTPAccrualClient create HTTPAccrualClient
func NewHTTPAccrualClient(URL string, poolSize int, poolTimeout time.Duration) *HTTPAccrualClient {
	url := URL
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	client := resty.
		New().
		SetBaseURL(url).
		SetTransport(&http.Transport{
			MaxConnsPerHost: poolSize,
			MaxIdleConns:    poolSize,
			IdleConnTimeout: poolTimeout,
		})
	return &HTTPAccrualClient{client: client}
}

// Close close client
func (c *HTTPAccrualClient) Close() error {
	return c.client.Close()
}

// GetOrder get orger from accrual service
func (c *HTTPAccrualClient) GetOrder(ctx context.Context, orderNumber string) (AccrualResponse, error) {
	var response AccrualResponse

	r, err := c.client.R().
		SetPathParam("orderNumber", orderNumber).
		SetResult(&response).
		Get("/api/orders/{orderNumber}")
	if err != nil {
		return AccrualResponse{}, err
	}
	if r.StatusCode() == http.StatusNoContent {
		return AccrualResponse{}, ErrOrderNotRegistered
	} else if r.StatusCode() == http.StatusTooManyRequests {
		return AccrualResponse{}, ErrTooManyRequest
	} else if r.StatusCode() != http.StatusOK {
		return AccrualResponse{}, fmt.Errorf("status code %d", r.StatusCode())
	}
	if err := validator.New().Struct(response); err != nil {
		return AccrualResponse{}, err
	}
	return response, nil
}
