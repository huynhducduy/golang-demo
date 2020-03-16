package app

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DB_HOST string
	DB_PORT string
	DB_USER string
	DB_PASS string
	DB_NAME string
	SECRET  string
}

var config Config

func readConfig() {

	viper.SetConfigFile(".env")

	viper.AddConfigPath("..")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
