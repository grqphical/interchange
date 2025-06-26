package main

import (
	"fmt"
	"log/slog"
	"net/http"

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

func buildHTTPRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if viper.GetBool("developmentMode") {
		r.Get("/debug", handlers.DebugHandler)
	}

	return r
}

func main() {
	logger := slog.New(&ApplicationLogHandler{})

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

	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("config changed, reloading config")
	})

	if !viper.GetBool("production") {
		logger.Warn("Interchange is running in development mode. Do not use this in production, instead pass the CLI flag '--production'")
		viper.WatchConfig()
	}

	logger.Info(fmt.Sprintf("Starting Interchange on %s:%d", viper.GetString("hostAddress"), viper.GetInt("port")))

	server := http.Server{
		Handler: buildHTTPRouter(),
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("hostAddress"), viper.GetInt("port")),
	}

	server.ListenAndServe()

}
