package net

import (
	"fmt"
	"syscall"

	"github.com/pedrogao/log"
)

type Poller struct {
	SocketFd int
	PollerFd int
}

func NewPoller(s *Socket) (*EventLoop, error) {
	kqueue, err := syscall.Kqueue()
	if err != nil {
		return nil, fmt.Errorf("create kqueue err: (%s)", err)
	}

	changeEvent := syscall.Kevent_t{
		Ident:  uint64(s.Fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
		Fflags: 0,
		Data:   0,
		Udata:  nil,
	}
	ret, err := syscall.Kevent(
		kqueue,
		[]syscall.Kevent_t{changeEvent},
		nil,
		nil,
	)
	if err != nil || ret == -1 {
		return nil, fmt.Errorf("register event to kqueue err: (%s)", err)
	}

	return &EventLoop{
		SocketFd: s.Fd,
		PollerFd: kqueue,
	}, nil
}

func (e *Poller) Handle(handler Handler) {
	for {
		log.Info("polling for new events...")
		newEvents := make([]syscall.Kevent_t, 10)
		numNewEvents, err := syscall.Kevent(e.PollerFd, nil, newEvents, nil)
		if err != nil {
			log.Infof("call kevent err: (%s)", err)
			continue
		}

		for i := 0; i < numNewEvents; i++ {
			currentEvent := newEvents[i]
			eventFD := int(currentEvent.Ident)

			if currentEvent.Flags&syscall.EV_EOF != 0 {
				log.Info("client disconnected.")
				_ = syscall.Close(eventFD)
			} else if eventFD == e.SocketFd {
				clientFD, clientAddr, err := syscall.Accept(eventFD)
				if err != nil {
					log.Info("failed to create Socket for connecting to client:", err)
					continue
				}
				log.Info("accepted new connection ", clientFD, " from ", eventFD, " addr ", clientAddr)

				socketEvent := syscall.Kevent_t{
					Ident:  uint64(clientFD),
					Filter: syscall.EVFILT_READ, // read
					Flags:  syscall.EV_ADD,
					Fflags: 0,
					Data:   0,
					Udata:  nil,
				}
				ret, err := syscall.Kevent(e.PollerFd, []syscall.Kevent_t{socketEvent}, nil, nil)
				if err != nil || ret == -1 {
					log.Infof("failed to register Socket event:", err)
					continue
				}
			} else if currentEvent.Filter&syscall.EVFILT_READ != 0 {
				clientSocket := &Socket{Fd: eventFD}
				handler(clientSocket)
			}
			// Ignore any other events
		}
	}
}
