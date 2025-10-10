package routematcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

type SSRPayload interface {
	AsMap() map[string]any
}

type MapPayload map[string]any

func (m MapPayload) AsMap() map[string]any {
	return m
}

type HandlerFunc[T SSRPayload] func(context.Context, map[string]string, url.Values) (T, error)

type Route struct {
	Pattern string
	Handler func(context.Context, map[string]string, url.Values) (SSRPayload, error)
}

func RouteOf[T SSRPayload](pattern string, handler HandlerFunc[T]) Route {
	return Route{
		Pattern: pattern,
		Handler: func(ctx context.Context, params map[string]string, query url.Values) (SSRPayload, error) {
			return handler(ctx, params, query)
		},
	}
}

type Matcher struct {
	patterns []compiledRoute
}

type compiledRoute struct {
	pattern Route
	regex   *regexp.Regexp
	params  []string
}

func New(routes []Route) *Matcher {
	compiled := make([]compiledRoute, 0, len(routes))
	for _, r := range routes {
		if r.Handler == nil {
			continue
		}
		regex, params := compilePattern(r.Pattern)
		compiled = append(compiled, compiledRoute{pattern: r, regex: regex, params: params})
	}

	return &Matcher{patterns: compiled}
}

func (m *Matcher) Fetch(ctx context.Context, req *http.Request) (SSRPayload, error) {
	cleanPath := path.Clean(req.URL.Path)
	query := req.URL.Query()
	for _, route := range m.patterns {
		matches := route.regex.FindStringSubmatch(cleanPath)
		if len(matches) == 0 {
			continue
		}

		params := make(map[string]string, len(route.params))
		for idx, name := range route.params {
			params[name] = matches[idx+1]
		}

		return route.pattern.Handler(ctx, params, query)
	}

	return nil, nil
}

func compilePattern(pattern string) (*regexp.Regexp, []string) {
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	paramNames := []string{}

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			name := strings.TrimPrefix(segment, ":")
			paramNames = append(paramNames, name)
			segments[i] = "([^/]+)"
		}
	}

	regexPattern := fmt.Sprintf("^/%s$", strings.Join(segments, "/"))
	return regexp.MustCompile(regexPattern), paramNames
}
