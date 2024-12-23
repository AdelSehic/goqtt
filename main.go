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

	"github.com/sevlyar/go-daemon"
)

func main() {
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		panic(err)
	}
	logger.Init(config.Logger)

	cntxt := &daemon.Context{
		PidFileName: "goqtt.pid",
		PidFilePerm: 0644,
		LogFileName: "goqtt.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"goqtt"},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		logger.Console.Fatal().Err(err).Msg("Failed to daemonize")
	}
	if d != nil {
		logger.Console.Info().Msg("Daemon process started successfully!")
		return
	}
	defer cntxt.Release()

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
		logger.Console.Info().Msg("Shutting down server")
		srv.Stop()
		workers.GlobalPool.StopWorkers()
		os.Exit(0)
	}()

	srv.Start()
}
