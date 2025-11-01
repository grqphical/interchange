package middleware

import (
	"net/http"
	"strings"

	"github.com/grqphical/interchange/templates"
	"github.com/spf13/viper"
)

func WhitelistMiddleware(next http.Handler) http.Handler {
	whitelist := viper.GetStringSlice("whitelist")

	fn := func(w http.ResponseWriter, r *http.Request) {
		if len(whitelist) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		var isValidIP = false
		for _, listIP := range whitelist {
			remoteIPAndPort := r.RemoteAddr
			remoteIP := strings.Split(remoteIPAndPort, ":")[0]

			if remoteIP == listIP {
				isValidIP = true
			}
		}

		if !isValidIP {
			templates.WriteError(w, http.StatusForbidden, "Forbidden")
		} else {
			next.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(fn)
}
