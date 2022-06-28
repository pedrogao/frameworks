package main

import (
	"bufio"
	"strings"

	"github.com/pedrogao/log"
	"github.com/pedrogao/net"
)

func main() {
	s, err := net.Listen("127.0.0.1", 8080)
	if err != nil {
		log.Fatalf("failed to create Socket:", err)
	}

	eventLoop, err := net.NewEventLoop(s)
	if err != nil {
		log.Fatalf("failed to create event loop:", err)
	}

	log.Info("server started. Waiting for incoming connections. ^C to exit.")

	eventLoop.Handle(func(s *net.Socket) {
		reader := bufio.NewReader(s)
		for {
			line, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(line) == "" {
				break
			}
			s.Write([]byte(line))
		}
		s.Close()
	})
}
