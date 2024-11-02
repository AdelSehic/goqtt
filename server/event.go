package server

import (
	"fmt"
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
	RootEvent.mtx.Lock()
	defer RootEvent.mtx.Unlock()

	evIterator := sliceiterator.NewIterator(strings.Split(job.EventString, "/"))
	event := RootEvent.subscribeHelper(evIterator, make([]*Connection, 0))
	event.Subscribers[job.Conn.ID] = job.Conn
	RootEvent.recursivePrint("")
}

func (ev *Event) subscribeHelper(it *sliceiterator.SliceIter[string], subs []*Connection) *Event {
	if it.IsLast() {
		return ev
	}

	level := it.Value()
	if _, found := ev.SubEvents[level]; !found {
		ev.SubEvents[level] = NewEvent(level)
	}

	return ev.SubEvents[level].subscribeHelper(it.Next(), subs)
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
	if it.IsLast() {
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
