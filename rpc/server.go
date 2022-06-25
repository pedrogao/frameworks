package rpc

import (
	"fmt"
	"io"
	"net"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/pedrogao/log"
	"github.com/pedrogao/rpc/codec"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodeType    codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodeType:    codec.JsonType,
}

// Server of rpc
type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("accept conn err: %s", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	var opt Option
	if err := jsoniter.NewDecoder(conn).Decode(&opt); err != nil {
		log.Errorf("decode option err: %s", err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Error("invalid packed, magic number in-correct")
		return
	}

	f := codec.NewCodecFuncMap[opt.CodeType]
	if f == nil {
		log.Error("invalid codec type")
		return
	}
	s.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := sync.Mutex{}
	wg := sync.WaitGroup{}
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, &sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, &sending, &wg)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h     *codec.Header
	argv  any
	reply any
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadeHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Errorf("read header err: %s", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	if err = cc.ReadBody(&req.argv); err != nil {
		log.Errorf("read request argv err: %s", err)
	}

	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header,
	body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(h, body); err != nil {
		log.Errorf("write response err: %s", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request,
	sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Infof("handle request: %+v, %+v", req.h, req.argv)
	req.reply = fmt.Sprintf("resp: %d", req.h.Seq)
	s.sendResponse(cc, req.h, req.reply, sending)
}

var DefaultServer = NewServer()

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}
