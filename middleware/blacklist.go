package middleware

import (
	"net/http"
	"strings"

	"github.com/grqphical/interchange/templates"
	"github.com/spf13/viper"
)

func BlacklistMiddleware(next http.Handler) http.Handler {
	blacklist := viper.GetStringSlice("blacklist")

	fn := func(w http.ResponseWriter, r *http.Request) {
		if len(blacklist) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		for _, listIP := range blacklist {
			remoteIPAndPort := r.RemoteAddr
			remoteIP := strings.Split(remoteIPAndPort, ":")[0]

			if remoteIP == listIP {
				templates.WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
