package main

import (
	"log"

	"github.com/spf13/viper"
)

func setDefaultConfig() {
	viper.SetDefault("port", 80)
	viper.SetDefault("hostAddress", "0.0.0.0")
}

func main() {
	setDefaultConfig()

	viper.SetConfigName("interchange")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/interchange")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Warning: No interchange.toml file found, reverting to default settings")
	}

}
