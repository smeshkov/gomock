package app

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/smeshkov/gomock/config"
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
	config.Log.Debug("proxying call to %s %s", req.Method, addr)

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
