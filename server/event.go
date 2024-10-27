package server

import (
	"fmt"
	"goqtt/logger"
	"strings"
	"sync"
)

const (
	EV_SUBSCRIBE = "SUB"
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
	fmt.Println(evName)
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
