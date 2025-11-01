package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/grqphical/interchange/handlers"
	"github.com/grqphical/interchange/middleware"
	"github.com/grqphical/interchange/templates"
	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

var Version string = "interchange/0.1.0"

// sets the default global configuration
func setDefaultConfig() {
	viper.SetDefault("port", 80)
	viper.SetDefault("hostAddress", "0.0.0.0")
	viper.SetDefault("developmentMode", true)
}

// build a new HTTP router to be used by interchange, creating the debug handlers if developmentMode is true
// and routing all the services defined in `interchange.toml`
func buildHTTPRouter(logger *ApplicationLogHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.BlacklistMiddleware)
	r.Use(chimiddleware.Logger)
	r.Use(middleware.WhitelistMiddleware)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		templates.WriteError(w, http.StatusNotFound, "Not Found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		templates.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	})

	if viper.GetBool("developmentMode") {
		r.Get("/debug", handlers.DebugHandler)
		r.Get("/debug/log", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
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
			proxy, success := handlers.BuildReverseProxyService(service, name)
			if !success {
				continue serviceLoop
			}

			route := path.Join(route.(string), "*")

			r.Handle(route, proxy)
		case "staticFS":
			routeStr := route.(string)

			if !strings.HasSuffix(routeStr, "/") {
				routeStr += "/"
			}

			fs, success := handlers.BuildStaticFileSystemHandler(service, name, routeStr)
			if !success {
				continue
			}

			if !strings.HasSuffix(routeStr, "*") {
				routeStr += "*"
			}

			r.Handle(routeStr, fs)

		default:
			slog.Error("ConfigurationError", "err", fmt.Sprintf("invalid mode set on service '%s'", name))

		}
		slog.Info(fmt.Sprintf("loaded service '%s' of type '%s'", name, serviceType))
	}

	return r
}

// starts a new instance of the server on a new thread
func startServer() *http.Server {
	server := http.Server{
		Handler: buildHTTPRouter(slog.Default().Handler().(*ApplicationLogHandler)),
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("hostAddress"), viper.GetInt("port")),
	}

	go func() {
		slog.Info(fmt.Sprintf("Starting Interchange on %s:%d", viper.GetString("hostAddress"), viper.GetInt("port")))
		if viper.Get("https") != nil {
			certFile := viper.GetString("https.certificate_file")
			if certFile == "" {
				slog.Error("Failed to initialize HTTPS: certificate_file not specified")
				return
			}

			keyFile := viper.GetString("https.key_file")
			if keyFile == "" {
				slog.Error("Failed to initialize HTTPS: key_file not specified")
				return
			}

			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("Failed to start server", "err", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Failed to start server", "err", err)
			}
		}

	}()

	return &server
}

func main() {
	logger := slog.New(&ApplicationLogHandler{})
	slog.SetDefault(logger)

	prod := flag.Bool("production", false, "Sets the reverse proxy to run in production mode disabling things such as config reloading")
	version := flag.Bool("version", false, "Prints the version")

	flag.Parse()

	if *version {
		fmt.Printf("%s\n", Version)
		return
	}

	if *prod {
		viper.Set("developmentMode", false)
	}

	setDefaultConfig()

	viper.SetConfigName("interchange")
	viper.SetConfigType("toml")
	// viper.AddConfigPath("$HOME/.config/interchange")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Warn("no interchange.toml found, using default configuration")
	}

	server := startServer()

	// restart the server if the configuration is reloaded, ensuring the old server shuts down gracefully first
	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("config changed, reloading config")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown server", "err", err)
			return
		}
		server = startServer()
	})

	if viper.GetBool("developmentMode") {
		logger.Warn("Interchange is running in development mode. Do not use this in production, instead pass the CLI flag '--production'")
		viper.WatchConfig()
	}

	// handle Ctrl+C, Ctrl+D signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", "err", err)
	}
}
