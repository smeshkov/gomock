package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Proxy wraps a reverse proxy with URL rewriting.
type Proxy struct {
	host   *url.URL
	target *url.URL
	proxy  *httputil.ReverseProxy
	log    *zap.Logger
}

var (
	errParseHostURL  = errors.New("error in parsing host URL")
	errParseProxyURL = errors.New("error in parsing proxy URL")
)

const (
	redirectMinCode = 300
	redirectMaxCode = 400
)

func newProxy(serverAddr, target string, log *zap.Logger) (*Proxy, error) {
	hostURL, err := url.Parse("http://localhost" + serverAddr)
	if err != nil {
		return nil, fmt.Errorf("%w [%s]", errParseHostURL, serverAddr)
	}

	proxyURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w [%s]", errParseProxyURL, target)
	}

	return &Proxy{
		host:   hostURL,
		target: proxyURL,
		proxy:  httputil.NewSingleHostReverseProxy(proxyURL),
		log:    log,
	}, nil
}

func (p *Proxy) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	wrapper := &responseWriterWrapper{
		ResponseWriter: writer,
		statusCode:     http.StatusOK,
	}

	// ---> request handling
	newReqQuery, err := p.adjustRedirectQuery(req.URL)
	if err != nil {
		p.log.Sugar().Errorf("error in adjusting request URI redirect location: %v", err)

		return
	}

	req.URL.RawQuery = newReqQuery

	p.log.Sugar().Debugf("proxying call to [%s]", req.RequestURI)
	p.proxy.ServeHTTP(wrapper, req)

	// ---> response handling
	if wrapper.statusCode < redirectMinCode || wrapper.statusCode >= redirectMaxCode {
		return
	}

	// handle redirect url
	location := writer.Header().Get("Location")
	if location != "" {
		redirect := p.adjustRedirectURL(location)
		p.log.Sugar().Debugf("redirect location [%s]", redirect)
		http.Redirect(writer, req, redirect, wrapper.statusCode)
	}
}

func (p *Proxy) adjustRedirectURL(location string) string {
	targetAddr := p.target.String()
	hostAddr := p.host.String()
	location = strings.ReplaceAll(location, targetAddr, hostAddr)

	before, after, found := strings.Cut(location, "?")
	if found {
		query, _ := url.QueryUnescape(after)
		query = url.QueryEscape(strings.ReplaceAll(query, targetAddr, hostAddr))

		return before + "?" + query
	}

	return location
}

func (p *Proxy) adjustRedirectQuery(reqURL *url.URL) (string, error) {
	query, err := url.QueryUnescape(reqURL.RawQuery)
	if err != nil {
		return "", fmt.Errorf("error in decoding query part of the location: %w", err)
	}

	return strings.ReplaceAll(query, p.target.String(), p.host.String()), nil
}
