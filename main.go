package main

import (
	"fmt"
	"goqtt/config"
	"goqtt/logger"
	"goqtt/server"
	"goqtt/workers"
	"runtime"
)

type TestJob struct {
	Id int
}

func (t *TestJob) Run() {
	fmt.Println(t.Id)
}

func main() {
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		panic(err)
	}
	logger.Init(config.Logger)

	srv := server.NewServer(config.Connector)
	if srv == nil {
		logger.HTTP.Panic().Err(err).Msg("Couldn't create a server")
	}
	// srv.Start()
	pool := workers.NewPool(config.Pool)
	pool.StartWorkers(runtime.NumCPU())

	for i := 0; i < 5; i++ {
		t := &TestJob{i*2}
		pool.QueueJob(t)
	}

	pool.StopWorkers()

	for i := 0; i < 5; i++ {
		t := &TestJob{i*2}
		pool.QueueJob(t)
	}
}
