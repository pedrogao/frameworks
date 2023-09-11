package net

import (
	"fmt"
	"sync/atomic"

	"github.com/pedrogao/log"
)

type Handler = func(s *Socket)

type EventLoop struct {
	poller *Poller
	quit   int32
}

func NewEventLoop(s *Socket) (*EventLoop, error) {
	poller, err := NewPoller()
	if err != nil {
		return nil, fmt.Errorf("create poller err: (%s)", err)
	}

	poller.RegisterAcceptHandler(s.Fd, func(s *Socket) {
		log.Info("handle accept: %s", s)
	})

	poller.RegisterErrorHandler(func(s *Socket) {
		log.Error("handle error: %s", s)
	})

	poller.RegisterWriteHandler(func(s *Socket) {
		log.Error("handle write: %s", s)
	})

	return &EventLoop{
		poller: poller,
	}, nil
}

func (e *EventLoop) Loop() {
	for {
		if atomic.LoadInt32(&e.quit) == 1 {
			break
		}

		e.poller.Poll()
	}
}

func (e *EventLoop) Handle(handle Handler) {
	e.poller.RegisterReadHandler(handle)
}

func (e *EventLoop) Quit() {
	log.Info("event loop quit")
	atomic.StoreInt32(&e.quit, 1)
}
