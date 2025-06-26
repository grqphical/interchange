package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

func setDefaultConfig() {
	viper.SetDefault("port", 80)
	viper.SetDefault("hostAddress", "0.0.0.0")
	viper.SetDefault("developmentMode", true)
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
		viper.WatchConfig()
	}

	log.Printf("Info: Starting Interchange on %s:%d\n", viper.GetString("hostAddress"), viper.GetInt("port"))

	server := http.Server{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("hostAddress"), viper.GetInt("port")),
	}

	server.ListenAndServe()

}
