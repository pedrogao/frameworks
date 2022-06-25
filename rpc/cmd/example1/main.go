package main

import (
	"fmt"
	"net"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pedrogao/log"
	"github.com/pedrogao/rpc"
	"github.com/pedrogao/rpc/codec"
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

	conn, _ := net.Dial("tcp", <-addr)
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)
	// write option
	_ = jsoniter.NewEncoder(conn).Encode(rpc.DefaultOption)
	// write data
	cc := codec.NewJsonCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("req: %d", h.Seq))

		_ = cc.ReadeHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Infof("reply: %s", reply)
	}
}
