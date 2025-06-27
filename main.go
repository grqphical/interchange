package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grqphical/interchange/handlers"
	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

func setDefaultConfig() {
	viper.SetDefault("port", 80)
	viper.SetDefault("hostAddress", "0.0.0.0")
	viper.SetDefault("developmentMode", true)
}

func buildHTTPRouter(logger *ApplicationLogHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if viper.GetBool("developmentMode") {
		r.Get("/debug", handlers.DebugHandler)
		r.Get("/debug/log", func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(logger.records)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}

serviceLoop:
	for name, service := range viper.GetStringMap("services") {
		service := service.(map[string]any)
		serviceType, exists := service["mode"]
		if !exists {
			slog.Error("ConfigurationError", "err", fmt.Sprintf("mode not set on service '%s'", name))
			continue serviceLoop
		}

		route, exists := service["route"]
		if !exists {
			slog.Error("ConfigurationError", "err", fmt.Sprintf("route not set on service '%s'", name))
			continue serviceLoop
		}

		switch serviceType.(string) {
		case "reverseProxy":
			target, exists := service["target"]
			if !exists {
				slog.Error("ConfigurationError", "err", fmt.Sprintf("target not set on service '%s'", name))
				continue serviceLoop
			}

			targetURL, err := url.ParseRequestURI(target.(string))
			if err != nil {
				slog.Error("ConfigurationError", "err", fmt.Sprintf("target is invalid URL on service '%s'", name))
				continue serviceLoop
			}

			proxy := &httputil.ReverseProxy{
				Rewrite: func(r *httputil.ProxyRequest) {
					r.SetURL(targetURL)

					r.SetXForwarded()
				},
			}

			r.Handle(route.(string), proxy)
			slog.Info(fmt.Sprintf("loaded service '%s' of type '%s'", name, serviceType))

		default:
			slog.Error("ConfigurationError", "err", fmt.Sprintf("invalid mode set on service '%s'"))

		}
	}

	return r
}

func startServer() *http.Server {
	server := http.Server{
		Handler: buildHTTPRouter(slog.Default().Handler().(*ApplicationLogHandler)),
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("hostAddress"), viper.GetInt("port")),
	}

	go func() {
		slog.Info(fmt.Sprintf("Starting Interchange on %s:%d", viper.GetString("hostAddress"), viper.GetInt("port")))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "err", err)
		}
	}()

	return &server
}

func main() {
	logger := slog.New(&ApplicationLogHandler{})
	slog.SetDefault(logger)

	flag.Bool("production", false, "Sets the reverse proxy to run in production mode disabling things such as config reloading")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	setDefaultConfig()

	viper.SetConfigName("interchange")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/interchange")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Warn("no interchange.toml found, using default configuration")
	}

	server := startServer()

	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("config changed, reloading config")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown server", "err", err)
		}
		server = startServer()
	})

	if !viper.GetBool("production") {
		logger.Warn("Interchange is running in development mode. Do not use this in production, instead pass the CLI flag '--production'")
		viper.WatchConfig()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", "err", err)
	}
}
