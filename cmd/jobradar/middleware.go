package main

import (
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"time"
)

// Do not log raw query strings: OAuth and onboarding contain one-time tokens.
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, wrapped.Status(), time.Since(start))
		}()
		next.ServeHTTP(wrapped, r)
	})
}
func privateNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
