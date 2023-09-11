package net

import (
	"fmt"
	"syscall"

	"github.com/pedrogao/log"
)

type Poller struct {
	PollerFd      int
	acceptFd      int
	readHandler   Handler
	writeHandler  Handler
	acceptHandler Handler
	errorHandler  Handler
}

func NewPoller() (*Poller, error) {
	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("create epoll err: (%s)", err)
	}

	return &Poller{
		PollerFd: fd,
		acceptFd: -1,
	}, nil
}

func (p *Poller) Poll() error {
	events := make([]syscall.EpollEvent, 1)
	n, err := syscall.EpollWait(p.PollerFd, events, -1)
	if err != nil {
		return fmt.Errorf("epoll wait err: (%s)", err)
	}

	for i := 0; i < n; i++ {
		event := events[i]
		socket := &Socket{Fd: int(event.Fd)}
		if event.Fd == int32(p.PollerFd) {
			// Accept event
			p.handleAccept(socket)
		} else {
			if event.Events&syscall.EPOLLIN == 1 {
				// Read event
				p.handleRead(socket)
			} else if event.Events&syscall.EPOLLOUT == 1 {
				// Write event
				p.handleWrite(socket)
			} else if event.Events&syscall.EPOLLERR == 1 {
				// Error occurred
				p.handleError(socket)
			}
		}
	}

	return nil
}

func (p *Poller) handleAccept(s *Socket) {
	if p.acceptHandler != nil || p.acceptFd == -1 {
		return
	}
	// Accept new connection
	clientFD, clientAddr, err := syscall.Accept(s.Fd)
	if err != nil {
		log.Error("failed to create Socket for connecting to client:", err)
		return
	}
	log.Info("accepted new connection ", clientFD, " from ", s.Fd, " addr ", clientAddr)

	p.registerEvent(&syscall.EpollEvent{
		Fd:     int32(clientFD),
		Events: syscall.EPOLLIN | syscall.EPOLLPRI | syscall.EPOLLERR | syscall.EPOLLOUT, // read/write/error events
	})
	p.handleAccept(&Socket{Fd: clientFD})
}

func (p *Poller) handleRead(s *Socket) {
	if p.readHandler != nil {
		p.readHandler(s)
	}
}

func (p *Poller) handleWrite(s *Socket) {
	if p.writeHandler != nil {
		p.writeHandler(s)
	}
}

func (p *Poller) handleError(s *Socket) {
	if p.errorHandler != nil {
		p.errorHandler(s)
	}
}

func (p *Poller) RegisterWriteHandler(handler Handler) {
	p.writeHandler = handler
}

func (p *Poller) RegisterReadHandler(handler Handler) {
	p.readHandler = handler
}

func (p *Poller) RegisterAcceptHandler(acceptFd int, handler Handler) {
	p.acceptHandler = handler
	p.acceptFd = acceptFd
	p.registerEvent(&syscall.EpollEvent{Fd: int32(acceptFd), Events: syscall.EPOLLIN | syscall.EPOLLPRI | syscall.EPOLLERR})
}

func (p *Poller) RegisterErrorHandler(handler Handler) {
	p.errorHandler = handler
}

func (p *Poller) registerEvent(event *syscall.EpollEvent) {
	syscall.EpollCtl(p.PollerFd, syscall.EPOLL_CTL_ADD, int(event.Fd), event)
}
