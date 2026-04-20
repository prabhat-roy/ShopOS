package proxy

import (
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

func New(target, stripPrefix string, timeout time.Duration, log *zap.Logger) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatal("invalid proxy target", zap.String("target", target), zap.Error(err))
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	rp := &httputil.ReverseProxy{
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("upstream error", zap.String("target", target), zap.String("path", r.URL.Path), zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"upstream service unavailable"}`, http.StatusBadGateway)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if stripPrefix != "" {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, stripPrefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		r.URL.Host = u.Host
		r.URL.Scheme = u.Scheme
		r.Header.Set("X-Forwarded-Host", r.Host)

		if isWebSocket(r) {
			proxyWebSocket(w, r, u.Host, log)
			return
		}

		rp.ServeHTTP(w, r)
	})
}

func isWebSocket(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, upstreamHost string, log *zap.Logger) {
	upstreamConn, err := net.DialTimeout("tcp", upstreamHost, 10*time.Second)
	if err != nil {
		log.Error("websocket upstream dial failed", zap.String("host", upstreamHost), zap.Error(err))
		http.Error(w, `{"error":"websocket upstream unavailable"}`, http.StatusBadGateway)
		return
	}
	defer upstreamConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, `{"error":"websocket not supported"}`, http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Error("websocket hijack failed", zap.Error(err))
		return
	}
	defer clientConn.Close()

	if err := r.Write(upstreamConn); err != nil {
		log.Error("websocket request forward failed", zap.Error(err))
		return
	}

	done := make(chan struct{}, 2)
	go func() { io.Copy(upstreamConn, clientConn); done <- struct{}{} }()
	go func() { io.Copy(clientConn, upstreamConn); done <- struct{}{} }()
	<-done
}
