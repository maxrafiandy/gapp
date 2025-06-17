package middleware

import (
	"net/http"
	"strings"

	"scm/api/app"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

var DefaultCORSConfig = CORSConfig{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	AllowCredentials: false,
}

func CORS(config CORSConfig) app.MiddlewareFunc {
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	allowCreds := "false"
	if config.AllowCredentials {
		allowCreds = "true"
	}

	return func(next app.HandlerFunc) app.HandlerFunc {
		return func(ctx *app.Context) error {
			w := ctx.Writer()
			r := ctx.Request()

			origin := r.Header.Get("Origin")
			if origin != "" {
				if contains(config.AllowOrigins, "*") || contains(config.AllowOrigins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", allowCreds)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return nil
			}

			return next(ctx)
		}
	}
}

func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
