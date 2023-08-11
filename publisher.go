package simple_subpub

import (
	"fmt"
	"sync"
)

var (
	ErrorIdExist    = fmt.Errorf("subscriber id already exist")
	ErrorIdNotExist = fmt.Errorf("no subscriber has this id")
)

// Publisher can send message to all subscribers
// T can be any type that's safe to do shallow copy.
type Publisher[T any] struct {
	subscribers map[string]*Subscriber[T]
	lock        sync.RWMutex
}

// Subscribe will return a new Subscriber using provided id, duplicated id is not allowed
func (p *Publisher[T]) Subscribe(id string) (*Subscriber[T], error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subscribers[id]; ok {
		return nil, ErrorIdExist
	}
	tmpChan := make(chan T)
	sub := &Subscriber[T]{
		Channel:     tmpChan,
		sentChannel: tmpChan,
		id:          id,
		publisher:   p,
		invalid:     false,
	}
	p.subscribers[id] = sub

	return sub, nil
}

// Unsubscribe will delete the Subscriber with provided id, not existed id is not allowed.
// Once unsubscribed, id will be available again.
// The subscriber may still exist for a while, since publishWorker may still hold it.
func (p *Publisher[T]) Unsubscribe(id string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subscribers[id]; !ok {
		return ErrorIdNotExist
	}
	tmp := p.subscribers[id]
	delete(p.subscribers, id)
	tmp.invalid = true
	return nil
}

// Publish will send a message in type T to all subscribers' id at call time.
//
// It's async and return the number of subscribers.
// By distribute the real work to publishWorker, mutex hold time can be reduced.
func (p *Publisher[T]) Publish(msg T) int {
	subs := p.snapshot()
	go p.publishWorker(subs, msg)
	return len(subs)
}

// snapshot will return all subscribers' id at call time
func (p *Publisher[T]) snapshot() []*Subscriber[T] {
	p.lock.RLock()
	values := make([]*Subscriber[T], len(p.subscribers))

	i := 0
	for _, v := range p.subscribers {
		values[i] = v
		i++
	}
	p.lock.RUnlock()
	return values
}

// publishWorker is the real worker that do message sending.
//
// It will try to send messages to a subscriber, if blocked, try next one.
// Once a subscriber been sent, it will be kicked out of list.
// If a subscriber has been unsubscribed, it will also be kicked out of lists.
func (p *Publisher[T]) publishWorker(targets []*Subscriber[T], msg T) {
	for {
		tmpList := make([]*Subscriber[T], 0)
		i := 0

		for _, target := range targets {
			if target.invalid {
				continue
			}
			select {
			case target.sentChannel <- msg:
				// do nothing
			default:
				tmpList = append(tmpList, target)
				i++
			}
		}

		if len(tmpList) == 0 {
			break
		}

		targets = tmpList
	}
}

// Subscriber can be used to receive message from publisher
type Subscriber[T any] struct {
	Channel     <-chan T
	sentChannel chan<- T
	id          string
	publisher   *Publisher[T]
	invalid     bool
}

// Unsubscribe will call publisher's related unsubscribe function
func (s *Subscriber[T]) Unsubscribe() error {
	return s.publisher.Unsubscribe(s.id)
}

// NewPublisher will return a new functional publisher
func NewPublisher[T any]() *Publisher[T] {
	return &Publisher[T]{
		subscribers: make(map[string]*Subscriber[T]),
	}
}
