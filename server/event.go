package server

import (
	"fmt"
	"goqtt/logger"
	"goqtt/sliceiterator"
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
	RootEvent = NewEvent("root")
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

func FindSubs(evString string) []*Connection {
	eventsIterator := sliceiterator.NewIterator(strings.Split(evString, "/"))

	toNotify := make([]*Connection, 0)
	return RootEvent.recursiveFind(eventsIterator, toNotify)
}

func (ev *Event) recursiveFind(it *sliceiterator.SliceIter[string], subs []*Connection) []*Connection {
	logger.Console.Info().Msgf("ASDASDASD %s - %+v", ev.Name, ev.Subscribers)
	if it.IsLast() {
		a := sliceiterator.MapValuesToSlice(ev.Subscribers)
		logger.Console.Info().Msgf("ASDASDASDASD: %+v", a)
		return append(subs, sliceiterator.MapValuesToSlice(ev.Subscribers)...)
	}

	if _, found := ev.SubEvents[it.Value()]; !found {
		return subs
	}
	return ev.SubEvents[it.Value()].recursiveFind(it.Next(), subs)
}

type PublishJob struct {
	EventString string
	Data        string
	Conn        *Connection
}

func (job *PublishJob) Run() {
	subscribers := FindSubs(job.EventString)
	for _, client := range subscribers {
		workers.GlobalPool.QueueJob(NewWriteJob(client.Conn, []byte(job.Data)))
	}
}

func (job *PublishJob) Summary() string {
	return fmt.Sprintf("Publish event to [%s] from [%s]", job.EventString, job.Conn.ID)
}
