package main

import (
	"fmt"
	"goqtt/config"
	"goqtt/logger"
)

func main() {

	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		panic(err)
	}
	fmt.Println(config)

	logger.Console.Error().Msg("Hello world!")
	logger.HTTP.Error().Msg("ERROR")
}
