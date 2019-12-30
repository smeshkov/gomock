package app

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type client struct {
	c *http.Client
}

func newClient() *client {
	return &client{
		c: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *client) proxy(req *http.Request, url *url.URL) (*http.Response, error) {
	var sb strings.Builder
	sb.WriteString(url.String())
	if req.URL.Path != "" {
		sb.WriteString(req.URL.Path)
	}
	if req.URL.RawQuery != "" {
		sb.WriteString("?" + req.URL.RawQuery)
	}

	addr := sb.String()
	zap.L().Debug("proxying call", zap.String("method", req.Method), zap.String("address", addr))

	r, err := http.NewRequest(req.Method, addr, req.Body)
	if err != nil {
		return nil, err
	}

	for k, vs := range req.Header {
		for _, v := range vs {
			r.Header.Add(k, v)
		}
	}

	return c.c.Do(r)
}
