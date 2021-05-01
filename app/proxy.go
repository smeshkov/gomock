package app

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type Proxy struct {
	host   *url.URL
	target *url.URL
	proxy  *httputil.ReverseProxy
	log    *zap.Logger
}

func newProxy(serverAddr, target string, log *zap.Logger) (*Proxy, error) {
	hostURL, err := url.Parse("http://localhost" + serverAddr)
	if err != nil {
		return nil, fmt.Errorf("error in parsing host URL [%s]", serverAddr)
	}
	proxyURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("error in parsing proxy URL [%s]", target)
	}
	return &Proxy{
		host:   hostURL,
		target: proxyURL,
		proxy:  httputil.NewSingleHostReverseProxy(proxyURL),
		log:    log,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapper := &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// ---> request handling
	newReqQuery, err := p.adjustRedirectQuery(r.URL)
	if err != nil {
		p.log.Sugar().Errorf("error in adjusting request URI redirect location: %v", err)
		return
	}
	r.URL.RawQuery = url.QueryEscape(newReqQuery)

	p.log.Sugar().Debugf("proxying call to [%s]", r.RequestURI)
	p.proxy.ServeHTTP(wrapper, r)

	// ---> response handling

	if wrapper.statusCode < 300 || wrapper.statusCode >= 400 {
		return
	}

	// handle redirect url
	location := w.Header().Get("Location")
	if location != "" {
		redirect := p.adjustRedirectURL(location)
		p.log.Sugar().Debugf("redirect location [%s]", redirect)
		http.Redirect(w, r, redirect, wrapper.statusCode)
	}
}

func (p *Proxy) adjustRedirectURL(location string) string {
	targetAddr := p.target.String()
	hostAddr := p.host.String()
	location = strings.Replace(location, targetAddr, hostAddr, -1)
	if i := strings.Index(location, "?"); i != -1 {
		query, _ := url.QueryUnescape(location[i+1:])
		query = url.QueryEscape(strings.Replace(query, targetAddr, hostAddr, -1))
		return location[:i] + "?" + query
	}
	return location
}

func (p *Proxy) adjustRedirectQuery(u *url.URL) (string, error) {
	q, err := url.QueryUnescape(u.RawQuery)
	if err != nil {
		return "", fmt.Errorf("error in decoding query part of the location: %w", err)
	}
	return url.QueryEscape(strings.Replace(q, p.target.String(), p.host.String(), -1)), nil
}
