package codec

import (
	"bufio"
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/pedrogao/log"
)

type JsonCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *jsoniter.Decoder
	enc  *jsoniter.Encoder
}

var _ Codec = (*JsonCodec)(nil) // must implement Codec

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &JsonCodec{
		conn: conn,
		buf:  buf,
		dec:  jsoniter.NewDecoder(conn),
		enc:  jsoniter.NewEncoder(buf),
	}
}

func (c *JsonCodec) Close() error {
	return c.conn.Close()
}

func (c *JsonCodec) ReadeHeader(header *Header) error {
	return c.dec.Decode(header)
}

func (c *JsonCodec) ReadBody(body any) error {
	return c.dec.Decode(body)
}

func (c *JsonCodec) Write(header *Header, body any) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err = c.enc.Encode(header); err != nil {
		log.Errorf("rpc codec: json encoding header err: %s", err)
		return err
	}

	if err = c.enc.Encode(body); err != nil {
		log.Errorf("rpc codec: json encoding body err: %s", err)
		return err
	}

	return nil
}
