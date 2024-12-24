package main

import (
	"goqtt/config"
	"goqtt/logger"
	"goqtt/server"
	"goqtt/workers"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	confFile := os.Getenv("CONFIG_PATH")
	if confFile == "" {
		confFile = "config/config.json"
	}

	config, err := config.LoadConfig(confFile)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Default.Info().Msg("Shutting down server")
		srv.Stop()
		workers.GlobalPool.StopWorkers()
		os.Exit(0)
	}()

	srv.Start()
}
