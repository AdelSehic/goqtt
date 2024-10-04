package workers

import (
	"context"
	"goqtt/config"
	"goqtt/logger"
	"sync"
)

var GlobalPool *Pool

type Job interface {
	Run()
	Summary() string
}

type Worker struct {
	Id    int
	Queue chan Job
}

type Pool struct {
	open   bool
	Queue  chan Job
	Result chan any
	Error  chan error
	Ctx    context.Context
	Stop   context.CancelFunc
	Wg     *sync.WaitGroup
}

func NewPool(cfg *config.Pool) *Pool {
	ctx, stop := context.WithCancel(context.Background())
	return &Pool{
		Queue:  make(chan Job),
		Result: make(chan any),
		Error:  make(chan error),
		Ctx:    ctx,
		Stop:   stop,
		Wg:     &sync.WaitGroup{},
	}
}

func (pool *Pool) StartWorkers(size int) *Pool {
	for i := 0; i < size; i++ {
		pool.Wg.Add(1)
		go func(p *Pool) {
			for {
				select {
				case <-p.Ctx.Done():
					logger.Console.Info().Msgf("Stopping worker with id[ %d ]", i)
					p.Wg.Done()
					return
				case job := <-p.Queue:
					job.Run()
					logger.Console.Debug().Msgf("Worker[ %d ] finished job", i)
				default:
				}
			}
		}(pool)
		logger.Console.Info().Msgf("Stared worker with id[ %d ]", i)
	}
	pool.open = true
	return pool
}

func (pool *Pool) QueueJob(job Job) {
	if pool.open {
		pool.Queue <- job
	} else {
		logger.Console.Error().Msg("Trying to add jobs to a closed pool!")
	}
}

func (pool *Pool) StopWorkers() {
	pool.Stop()
	pool.open = false
}
