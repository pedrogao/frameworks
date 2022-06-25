package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pedrogao/log"
	"github.com/pedrogao/rpc"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatalf("network err: %s", err)
	}

	log.Infof("start rpc server on: %s", l.Addr())
	addr <- l.Addr().String()
	rpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	client, err := rpc.Dial("tcp", <-addr)
	if err != nil {
		log.Fatalf("rpc dial err: %s", err)
	}

	defer func() {
		client.Close()
	}()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			args := fmt.Sprintf("rpc req: %d", i)
			var reply string
			if err = client.Call("Foo.Sum", args, &reply); err != nil {
				log.Errorf("call Foo.Sum err: %s", err)
			} else {
				log.Infof("reply: %s", reply)
			}
		}(i)
	}
	wg.Wait()
}
