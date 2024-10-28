package server

import (
	"fmt"
	"goqtt/logger"
	"goqtt/workers"
	"strings"
	"sync"
)

const (
	EV_SUBSCRIBE = "SUB"
	EV_PUBLISH   = "PUB"
)

type Event struct {
	Name        string
	Subscribers map[string]*Connection
	SubEvents   map[string]*Event
	mtx         *sync.Mutex
}

var RootEvent *Event

func PrintAllEvents() {
	RootEvent.recursivePrint("root")
}

func (ev *Event) recursivePrint(evName string) {
	evName += "/" + ev.Name
	clients := make([]string, 0, 10)
	for _, client := range ev.Subscribers {
		clients = append(clients, client.ID)
	}
	fmt.Println(evName, clients)
	for _, nextEv := range ev.SubEvents {
		nextEv.recursivePrint(evName)
	}
}

func NewEvent(name string) *Event {
	return &Event{
		Name:        name,
		Subscribers: make(map[string]*Connection),
		SubEvents:   make(map[string]*Event),
	}
}

func EventsInit() {
	RootEvent = NewEvent("")
	RootEvent.mtx = &sync.Mutex{}
}

type SubscribeJob struct {
	EventString string
	Conn        *Connection
}

func (job *SubscribeJob) Run() {
	events := strings.Split(job.EventString, "/")
	var nextEvent *Event
	RootEvent.mtx.Lock()

	nextEvent = RootEvent
	for _, evName := range events {
		if _, found := nextEvent.SubEvents[evName]; !found {
			nextEvent.SubEvents[evName] = NewEvent(evName)
		}
		nextEvent = nextEvent.SubEvents[evName]
	}
	nextEvent.Subscribers[job.Conn.ID] = job.Conn
	logger.Console.Info().Msgf("%+v\n", RootEvent)

	PrintAllEvents()
	RootEvent.mtx.Unlock()
}

func (job *SubscribeJob) Summary() string {
	return fmt.Sprintf("%s subscribed to event [%s]", job.Conn.ID, job.EventString)
}

func FindEvent(evString string) *Event {
	evPath := ""
	events := strings.Split(evString, "/")

	var nextEvent *Event
	nextEvent = RootEvent
	for _, evName := range events {
		if _, found := nextEvent.SubEvents[evName]; !found {
			return nil
		}
		nextEvent = nextEvent.SubEvents[evName]
		evPath += "/" + nextEvent.Name
	}
	logger.Console.Info().Msgf("Found event [%s]", evPath)
	return nextEvent
}

type PublishJob struct {
	EventString string
	Data        string
	Conn        *Connection
}

func (job *PublishJob) Run() {
	ev := FindEvent(job.EventString)
	if ev == nil {
		return
	}
	for _, client := range ev.Subscribers {
		workers.GlobalPool.QueueJob(NewWriteJob(client.Conn, []byte(job.Data)))
	}
}

func (job *PublishJob) Summary() string {
	return fmt.Sprintf("Publish event to [%s] from [%s]", job.EventString, job.Conn.ID)
}
