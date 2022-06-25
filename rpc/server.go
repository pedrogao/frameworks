package rpc

import (
	"fmt"
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
	sending sync.Mutex
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("accept conn err: %s", err)
			continue
		}
		// 一个新的连接
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
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
	cc := f(conn)
	defer cc.Close()

	s.serveCodec(cc)
}

func (s *Server) serveCodec(cc codec.Codec) {
	for {
		// 读取请求
		req, err := s.readRequest(cc)
		if err != nil {
			req.h.Error = err.Error()
			// err response
			s.sendResponse(cc, req.h, nil)
			continue
		}
		go s.handleRequest(cc, req)
	}
}

type request struct {
	h     *codec.Header
	argv  any
	reply any
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadeHeader(&h); err != nil {
		log.Errorf("read header err: %s", err)
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
		return nil, fmt.Errorf("read request body err: %s", err)
	}

	return req, nil
}

func (s *Server) handleRequest(cc codec.Codec, req *request) {
	log.Infof("handle request: %+v, %+v", req.h, req.argv)
	req.reply = fmt.Sprintf("resp: %d", req.h.Seq)
	s.sendResponse(cc, req.h, req.reply)
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body any) {
	s.sending.Lock()
	defer s.sending.Unlock()

	if err := cc.Write(h, body); err != nil {
		log.Errorf("write response err: %s", err)
	}
}

var DefaultServer = NewServer()

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}
