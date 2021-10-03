package main

import (
	"sync"
	"regexp"
	"context"
	"net/http"
)

type route struct {
	method string
	pattern *regexp.Regexp
	handler http.HandlerFunc
}

type Router struct {
	routes []route
	rm sync.RWMutex
}

type ctxPathParams struct{}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Search in routes for a pattern that matches the current request
	// cf. https://benhoyt.com/writings/go-routing/
	r.rm.RLock()
	defer r.rm.RUnlock()

	for _,rt := range r.routes {
		// Extract path params and add them to the request context
		pathMatches := rt.pattern.FindStringSubmatch(req.URL.Path)
		if len(pathMatches) > 0 && rt.method == req.Method {
			ctx := context.WithValue(req.Context(), ctxPathParams{}, pathMatches[1:])
			rt.handler(w, req.WithContext(ctx))
			return
		}
	}
	http.NotFound(w, req)
}

func (r *Router) Get(urlPattern string, h http.HandlerFunc) {
	r.rm.Lock()
	defer r.rm.Unlock()
	r.routes = append(r.routes, route{"GET", regexp.MustCompile(urlPattern), h})
}

func (r *Router) Post(urlPattern string, h http.HandlerFunc) {
	r.rm.Lock()
	defer r.rm.Unlock()
	r.routes = append(r.routes, route{"POST", regexp.MustCompile(urlPattern), h})
}
