package main

import (
	"goqtt/config"
	"goqtt/logger"
	"goqtt/server"
	"goqtt/workers"
	"runtime"
)

func main() {
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		panic(err)
	}
	logger.Init(config.Logger)

	workers.GlobalPool = workers.NewPool(config.Pool)
	workers.GlobalPool.StartWorkers(runtime.NumCPU())

	srv := server.NewServer(config.Connector)
	if srv == nil {
		logger.HTTP.Panic().Err(err).Msg("Couldn't create a server")
	}

	srv.Start()
}
