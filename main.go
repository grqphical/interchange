package main

import (
	"fmt"
	"log"
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
		log.Println("Warning: No interchange.toml file found, reverting to default settings")
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("Info: config file changed, reloading config")
	})

	if !viper.GetBool("production") {
		log.Println("Warning: Interchange is running in development mode. Do not use this in production, instead pass the CLI flag '--production'")
		viper.WatchConfig()
	}

	log.Printf("Info: Starting Interchange on %s:%d\n", viper.GetString("hostAddress"), viper.GetInt("port"))

	server := http.Server{
		Handler: buildHTTPRouter(),
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("hostAddress"), viper.GetInt("port")),
	}

	server.ListenAndServe()

}
