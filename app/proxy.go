package app

import (
	"net/http"
	"net/url"
	"time"
)

type client struct {
	c *http.Client
}

func newClient() *client {
	return &client{
		c: &http.Client{
			// Transport: transport,
			Timeout: 120 * time.Second,
		},
	}
}

func (c *client) proxy(req *http.Request, url *url.URL) (*http.Response, error) {
	req.URL = url
	return c.c.Do(req)
}
