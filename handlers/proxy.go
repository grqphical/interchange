package handlers

import (
	"fmt"
	"log/slog"
	"net/http/httputil"
	"net/url"
)

func BuildReverseProxyService(service map[string]any, name string) (*httputil.ReverseProxy, bool) {
	target, exists := service["target"]
	if !exists {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("target not set on service '%s'", name))
		return nil, false
	}

	targetURL, err := url.ParseRequestURI(target.(string))
	if err != nil {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("target is invalid URL on service '%s'", name))
		return nil, false
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(targetURL)

			r.SetXForwarded()

			r.Out.Header.Set("Via", fmt.Sprintf("%s interchange", r.In.Proto))
		},
	}

	return proxy, true
}
