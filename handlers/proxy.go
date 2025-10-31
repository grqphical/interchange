package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/grqphical/interchange/templates"
)

// creates a new reverse proxy service based on the given user configuration
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

	forwardErrors, exists := service["forwarderrors"]
	if !exists {
		forwardErrors = false
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode >= 400 && !forwardErrors.(bool) {
			r.Body.Close()
			var buf bytes.Buffer
			templates.WriteError(&buf, r.StatusCode, http.StatusText(r.StatusCode))
			r.Body = io.NopCloser(&buf)
			r.Header.Set("Content-Type", "text/html")
			r.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
		}
		return nil
	}

	return proxy, true
}
