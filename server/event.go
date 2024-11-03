package server

import (
	"fmt"
	"goqtt/sliceiterator"
	"goqtt/workers"
	"strings"
	"sync"
)

const (
	EV_SUBSCRIBE   = "SUB"
	EV_PUBLISH     = "PUB"
	EV_ACKNOWLEDGE = "ACK"
)

type Event struct {
	Name        string
	Parent      *Event
	Subscribers map[string]*Connection
	SubEvents   map[string]*Event
	allWildcard map[string]*Connection
	mtx         *sync.Mutex
}

var RootEvent *Event

func PrintAllEvents() {
	RootEvent.recursivePrint("root")
}

func (ev *Event) recursivePrint(evName string) {
	evName += "/" + ev.Name
	clients := make([]string, 0, 10)
	wcClients := make([]string, 0, 10)
	for _, client := range ev.Subscribers {
		clients = append(clients, client.ID)
	}
	for _, wc := range ev.allWildcard {
		wcClients = append(wcClients, wc.ID)
	}
	fmt.Println(evName, clients, wcClients)
	for _, nextEv := range ev.SubEvents {
		nextEv.recursivePrint(evName)
	}
}

func NewEvent(name string, parent *Event) *Event {
	return &Event{
		Name:        name,
		Parent:      parent,
		Subscribers: make(map[string]*Connection),
		allWildcard: make(map[string]*Connection),
		SubEvents:   make(map[string]*Event),
	}
}

func EventsInit() {
	RootEvent = NewEvent("root", nil)
	RootEvent.Parent = RootEvent
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

	RootEvent.subscribeHelper(evIterator, job.Conn)
	RootEvent.recursivePrint("")
}

func (ev *Event) subscribeHelper(it *sliceiterator.SliceIter[string], sub *Connection) {
	if it.IsLast() {
		ev.Subscribers[sub.ID] = sub
		return
	}
	level := it.Value()

	if level == "#" {
		ev.allWildcard[sub.ID] = sub
		return
	}

	if _, found := ev.SubEvents[level]; !found {
		ev.SubEvents[level] = NewEvent(level, ev)
	}

	// recursive call
	ev.SubEvents[level].subscribeHelper(it.Next(), sub)
}

func (job *SubscribeJob) Summary() string {
	return fmt.Sprintf("%s subscribed to event [%s]", job.Conn.ID, job.EventString)
}

func FindSubs(evString string) map[string]*Connection {
	eventsIterator := sliceiterator.NewIterator(strings.Split(evString, "/"))

	toNotify := make(map[string]*Connection)
	return RootEvent.recursiveFind(eventsIterator, toNotify)
}

func (ev *Event) recursiveFind(it *sliceiterator.SliceIter[string], subs map[string]*Connection) map[string]*Connection {
	for k, v := range ev.allWildcard {
		subs[k] = v
	}
	if it.IsLast() {
		for k, v := range ev.Subscribers {
			subs[k] = v
		}
		return subs
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
	QoS         int8
}

func (job *PublishJob) Run() {
	subscribers := FindSubs(job.EventString)
	for _, client := range subscribers {
		workers.GlobalPool.QueueJob(NewWriteJob(client, []byte(job.Data), job.QoS))
	}
}

func (job *PublishJob) Summary() string {
	return fmt.Sprintf("Publish event to [%s] from [%s]", job.EventString, job.Conn.ID)
}
