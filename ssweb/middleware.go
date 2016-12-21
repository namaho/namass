package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

func ValidateServerToken(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := ctx.Value("config").(Config)
		log.Debug("validating server token: " + r.Header.Get("ServerToken"))
		if config.ServerToken != r.Header.Get("ServerToken") {
			w.WriteHeader(http.StatusForbidden)
			log.Warn("mismatching server token")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
