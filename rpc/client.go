package rpc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/pedrogao/log"
	"github.com/pedrogao/rpc/codec"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          any
	Reply         any
	Error         error
	Done          chan *Call // notify call is done
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call // 已发出的请求
	closing  bool
	shutdown bool
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closing {
		return ErrShutdown
	}

	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.shutdown && !c.closing
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}

	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()

	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

func (c *Client) terminateCalls(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sending.Lock()
	defer c.sending.Unlock()

	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

func (c *Client) receive() {
	var err error

	for err == nil {
		h := codec.Header{}
		if err = c.cc.ReadeHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		// switch 骚操作
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			if err = c.cc.ReadBody(&call.Reply); err != nil {
				call.Error = fmt.Errorf("read body err: %s", err)
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

func (c *Client) send(call *Call) {
	// 如果是发送请求，那么 sending 加锁
	c.sending.Lock()
	defer c.sending.Unlock()

	// 注册调用
	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done() // 完成
		return
	}

	log.Infof("send call, %s - %d", call.ServiceMethod, call.Seq)
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	if err = c.cc.Write(&c.header, &call.Args); err != nil {
		log.Errorf("write call body err: %s", err)

		call = c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) Go(serviceMethod string, args, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 1)
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	// 发送请求
	c.send(call)

	return call
}

func (c *Client) Call(serviceMethod string, args, reply any) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done

	return call.Error
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodeType]
	if f == nil {
		log.Errorf("invalid code type: %s", opt.CodeType)
		return nil, fmt.Errorf("invalid code type: %s", opt.CodeType)
	}

	if err := jsoniter.NewEncoder(conn).Encode(opt); err != nil {
		log.Errorf("encode option err: %s", err)
		_ = conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     opt,
		seq:     1,
		pending: map[uint64]*Call{},
	}
	go client.receive() // 处理 call

	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, fmt.Errorf("number of options: %d is more than 1", len(opts))
	}

	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodeType == "" {
		opt.CodeType = DefaultOption.CodeType
	}
	return opt, nil
}

func Dial(network, address string, opts ...*Option) (*Client, error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("dial net err: %s", err)
	}

	return NewClient(conn, opt)
}
