package simulator

import (
	"fmt"
)

const (
	TaskBufferSize = 256
)

type Reconcile struct {
	Key    string
	Action string
	Object interface{}
}

type Simulator struct {
	storage map[string]interface{}
	tasks   chan task
	watcher map[string][]chan Reconcile
}

func New() *Simulator {
	return &Simulator{
		storage: make(map[string]interface{}),
		tasks:   make(chan task, TaskBufferSize),
		watcher: make(map[string][]chan Reconcile),
	}
}

func (s *Simulator) Run() {
	go func() {
		for t := range s.tasks {
			s.loop(t)
		}
	}()
}

func (s *Simulator) Get(key string) (interface{}, error) {
	result := <-s.send(msgGet{key: key})
	return result.value, result.err
}

func (s *Simulator) Set(key string, object interface{}) {
	<-s.send(msgSet{key: key, object: object})
}

func (s *Simulator) Delete(key string) error {
	result := <-s.send(msgDelete{key: key})
	return result.err
}

func (s *Simulator) Watch(key string, ch chan Reconcile) {
	<-s.send(msgWatch{key: key, ch: ch})
}

func (s *Simulator) send(msg interface{}) chan result {
	ch := make(chan result)
	s.tasks <- task{
		input:  msg,
		result: ch,
	}
	return ch
}

func (s *Simulator) loop(t task) {
	var (
		result  interface{}
		err     error
		changed *Reconcile
	)

	switch msg := t.input.(type) {
	case msgGet:
		result, err = s.get(msg.key)
	case msgSet:
		s.set(msg.key, msg.object)
		changed = &Reconcile{
			Key:    msg.key,
			Action: "set",
			Object: msg.object,
		}
		result, err = nil, nil
	case msgDelete:
		changed = &Reconcile{
			Key:    msg.key,
			Action: "delete",
		}
		result, err = nil, s.delete(msg.key)
	default:
		result, err = nil, fmt.Errorf("undefined message type")
	}

	t.response(result, err)
	if changed != nil {
		s.reconcile(*changed)
	}
}

func (s *Simulator) get(key string) (interface{}, error) {
	if object, exists := s.storage[key]; exists {
		return object, nil
	}
	return nil, fmt.Errorf("object with %v key does not exist", key)
}

func (s *Simulator) set(key string, object interface{}) {
	s.storage[key] = object
}

func (s *Simulator) delete(key string) error {
	if _, exists := s.storage[key]; !exists {
		return fmt.Errorf("object with %v key does not exist", key)
	}
	delete(s.storage, key)
	return nil
}

func (s *Simulator) watch(key string, ch chan Reconcile) {
	if _, exists := s.watcher[key]; !exists {
		s.watcher[key] = make([]chan Reconcile, 0)
	}
	s.watcher[key] = append(s.watcher[key], ch)
}

func (s *Simulator) reconcile(reconcile Reconcile) {
	channels, exists := s.watcher[reconcile.Key]
	if !exists {
		return
	}
	for _, ch := range channels {
		select {
		case ch <- reconcile:
		default:
		}
	}
}

type result struct {
	value interface{}
	err   error
}

type task struct {
	input  interface{}
	result chan result
}

func (t task) response(value interface{}, err error) {
	select {
	case t.result <- result{value: value, err: err}:
	default:
	}
}

type msgGet struct {
	key string
}

type msgSet struct {
	key    string
	object interface{}
}

type msgDelete struct {
	key string
}

type msgWatch struct {
	key string
	ch  chan Reconcile
}
