package packet

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/core/buffer"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
)

const (
	dataBit      = 0 << 7 // 数据标识
	heartbeatBit = 1 << 7 // 心跳标识
)

type NocopyReader interface {
	// Next returns a slice containing the next n bytes from the buffer,
	// advancing the buffer as if the bytes had been returned by Read.
	Next(n int) (p []byte, err error)

	// Peek returns the next n bytes without advancing the reader.
	Peek(n int) (buf []byte, err error)

	// Release the memory space occupied by all read slices.
	Release() (err error)

	Slice(n int) (r NocopyReader, err error)
}

type Packer interface {
	// ReadBuffer 以buffer的形式读取消息
	ReadBuffer(reader io.Reader) (buffer.Buffer, error)
	// PackBuffer 以buffer的形式打包消息
	PackBuffer(message *Message) (*buffer.NocopyBuffer, error)
	// ReadMessage 读取消息
	ReadMessage(reader io.Reader) ([]byte, error)
	// PackMessage 打包消息
	PackMessage(message *Message) ([]byte, error)
	// UnpackMessage 解包消息
	UnpackMessage(data []byte) (*Message, error)
	// PackHeartbeat 打包心跳
	PackHeartbeat() ([]byte, error)
	// CheckHeartbeat 检测心跳包
	CheckHeartbeat(data []byte) (bool, error)
}

type defaultPacker struct {
	opts       *options
	heartbeat  []byte
	bufferPool sync.Pool // 复用 bytes.Buffer
}

// NewPacker 创建一个新的消息打包器
func NewPacker(opts ...Option) (*defaultPacker, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.routeBytes != 1 && o.routeBytes != 2 && o.routeBytes != 4 {
		return nil, errors.NewWithCode(codes.InvalidArgument, "the number of route bytes must be 1, 2, or 4")
	}

	if o.seqBytes != 0 && o.seqBytes != 1 && o.seqBytes != 2 && o.seqBytes != 4 {
		return nil, errors.NewWithCode(codes.InvalidArgument, "the number of seq bytes must be 0, 1, 2, or 4")
	}

	if o.bufferBytes < 0 {
		return nil, errors.NewWithCode(codes.InvalidArgument, "the number of buffer bytes must be greater than or equal to 0")
	}

	return &defaultPacker{
		opts:      o,
		heartbeat: makeHeartbeat(o.byteOrder),
		bufferPool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, defaultSizeBytes+defaultHeaderBytes+o.routeBytes+o.seqBytes+256))
			},
		},
	}, nil
}

// MustNewPacker 创建一个新的消息打包器，失败时 panic
func MustNewPacker(opts ...Option) *defaultPacker {
	p, err := NewPacker(opts...)
	if err != nil {
		log.Fatalf("create packer failed: %v", err)
	}
	return p
}

// ReadBuffer 以buffer的形式读取消息
func (p *defaultPacker) ReadBuffer(reader io.Reader) (buffer.Buffer, error) {
	buf1 := buffer.MallocBytes(defaultSizeBytes)
	defer buf1.Release()

	if _, err := io.ReadFull(reader, buf1.Bytes()); err != nil {
		return nil, err
	}

	size := p.opts.byteOrder.Uint32(buf1.Bytes())

	if size == 0 {
		return nil, nil
	}

	buf2 := buffer.MallocBytes(int(defaultSizeBytes + size))
	data := buf2.Bytes()

	copy(data[:defaultSizeBytes], buf1.Bytes())

	if _, err := io.ReadFull(reader, data[defaultSizeBytes:]); err != nil {
		buf2.Release()
		return nil, err
	}

	return buf2, nil
}

// PackBuffer 以buffer的形式打包消息
func (p *defaultPacker) PackBuffer(message *Message) (*buffer.NocopyBuffer, error) {
	if err := p.validateMessage(message); err != nil {
		return nil, err
	}

	writer := buffer.MallocWriter(defaultSizeBytes + defaultHeaderBytes + p.opts.routeBytes + p.opts.seqBytes)
	writer.WriteInt32s(p.opts.byteOrder, int32(defaultHeaderBytes+p.opts.routeBytes+p.opts.seqBytes+len(message.Buffer)))
	writer.WriteInt8s(int8(dataBit))

	p.writeRouteToWriter(writer, message.Route)
	p.writeSeqToWriter(writer, message.Seq)

	return buffer.NewNocopyBuffer(writer, message.Buffer), nil
}

// ReadMessage 读取消息
func (p *defaultPacker) ReadMessage(reader io.Reader) ([]byte, error) {
	buf := make([]byte, defaultSizeBytes)

	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}

	size := p.opts.byteOrder.Uint32(buf)

	if size == 0 {
		return nil, nil
	}

	data := make([]byte, int(defaultSizeBytes+size))

	copy(data[:defaultSizeBytes], buf)

	if _, err := io.ReadFull(reader, data[defaultSizeBytes:]); err != nil {
		return nil, err
	}

	return data, nil
}

// nocopyReadMessage 无拷贝读取消息
func (p *defaultPacker) nocopyReadMessage(reader NocopyReader) ([]byte, error) {
	buf, err := reader.Peek(defaultSizeBytes)
	if err != nil {
		return nil, err
	}

	var size uint32

	if p.opts.byteOrder == binary.BigEndian {
		size = binary.BigEndian.Uint32(buf)
	} else {
		size = binary.LittleEndian.Uint32(buf)
	}

	if size == 0 {
		return nil, nil
	}

	n := int(defaultSizeBytes + size)

	r, err := reader.Slice(n)
	if err != nil {
		return nil, err
	}

	buf, err = r.Next(n)
	if err != nil {
		return nil, err
	}

	if err = reader.Release(); err != nil {
		return nil, err
	}

	return buf, nil
}

// PackMessage 打包消息
func (p *defaultPacker) PackMessage(message *Message) ([]byte, error) {
	if err := p.validateMessage(message); err != nil {
		return nil, err
	}

	size := defaultHeaderBytes + p.opts.routeBytes + p.opts.seqBytes + len(message.Buffer)
	buf := p.getBuffer()
	defer p.putBuffer(buf)

	buf.Grow(size + defaultSizeBytes)

	if err := binary.Write(buf, p.opts.byteOrder, int32(size)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, p.opts.byteOrder, int8(dataBit)); err != nil {
		return nil, err
	}

	if err := p.writeRouteToBuf(buf, message.Route); err != nil {
		return nil, err
	}

	if err := p.writeSeqToBuf(buf, message.Seq); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, p.opts.byteOrder, message.Buffer); err != nil {
		return nil, err
	}

	// 复制一份返回，因为 buf 会被放回池中复用
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
}

// UnpackMessage 解包消息
func (p *defaultPacker) UnpackMessage(data []byte) (*Message, error) {
	ln := defaultSizeBytes + defaultHeaderBytes + p.opts.routeBytes + p.opts.seqBytes

	if len(data)-ln < 0 {
		return nil, errors.ErrInvalidMessage
	}

	reader := bytes.NewReader(data)

	var size uint32
	if err := binary.Read(reader, p.opts.byteOrder, &size); err != nil {
		return nil, err
	}

	if uint64(len(data))-defaultSizeBytes != uint64(size) {
		return nil, errors.ErrInvalidMessage
	}

	var header uint8
	if err := binary.Read(reader, p.opts.byteOrder, &header); err != nil {
		return nil, err
	}

	if header&dataBit != dataBit {
		return nil, errors.ErrInvalidMessage
	}

	route, err := p.readRoute(reader)
	if err != nil {
		return nil, err
	}

	seq, err := p.readSeq(reader)
	if err != nil {
		return nil, err
	}

	return &Message{
		Route:  route,
		Seq:    seq,
		Buffer: data[ln:],
	}, nil
}

// PackHeartbeat 打包心跳
func (p *defaultPacker) PackHeartbeat() ([]byte, error) {
	if p.opts.heartbeatTime {
		buf := p.getBuffer()
		defer p.putBuffer(buf)

		size := defaultHeaderBytes + defaultHeartbeatTimeBytes
		buf.Grow(defaultSizeBytes + size)

		if err := binary.Write(buf, p.opts.byteOrder, uint32(size)); err != nil {
			return nil, err
		}

		if err := binary.Write(buf, p.opts.byteOrder, uint8(heartbeatBit)); err != nil {
			return nil, err
		}

		if err := binary.Write(buf, p.opts.byteOrder, time.Now().UnixNano()); err != nil {
			return nil, err
		}

		result := make([]byte, buf.Len())
		copy(result, buf.Bytes())

		return result, nil
	}

	return p.heartbeat, nil
}

// CheckHeartbeat 检测心跳包
func (p *defaultPacker) CheckHeartbeat(data []byte) (bool, error) {
	if len(data) < defaultSizeBytes+defaultHeaderBytes {
		return false, errors.ErrInvalidMessage
	}

	reader := bytes.NewReader(data)

	var size uint32
	if err := binary.Read(reader, p.opts.byteOrder, &size); err != nil {
		return false, err
	}

	if uint64(len(data))-defaultSizeBytes != uint64(size) {
		return false, errors.ErrInvalidMessage
	}

	var header uint8
	if err := binary.Read(reader, p.opts.byteOrder, &header); err != nil {
		return false, err
	}

	return header&heartbeatBit == heartbeatBit, nil
}

// ==================== 私有方法 ====================

// validateMessage 校验消息
func (p *defaultPacker) validateMessage(message *Message) error {
	// 校验路由范围
	maxRoute := int32(1<<(8*p.opts.routeBytes-1) - 1)
	minRoute := int32(-1 << (8*p.opts.routeBytes - 1))
	if message.Route > maxRoute || message.Route < minRoute {
		return errors.ErrRouteOverflow
	}

	// 校验序列号范围
	if p.opts.seqBytes > 0 {
		maxSeq := int32(1<<(8*p.opts.seqBytes-1) - 1)
		minSeq := int32(-1 << (8*p.opts.seqBytes - 1))
		if message.Seq > maxSeq || message.Seq < minSeq {
			return errors.ErrSeqOverflow
		}
	}

	// 校验消息大小
	if len(message.Buffer) > p.opts.bufferBytes {
		return errors.ErrMessageTooLarge
	}

	return nil
}

// writeRouteToWriter 写路由到 buffer.Writer
func (p *defaultPacker) writeRouteToWriter(writer *buffer.Writer, route int32) {
	switch p.opts.routeBytes {
	case 1:
		writer.WriteInt8s(int8(route))
	case 2:
		writer.WriteInt16s(p.opts.byteOrder, int16(route))
	case 4:
		writer.WriteInt32s(p.opts.byteOrder, route)
	}
}

// writeSeqToWriter 写序列号到 buffer.Writer
func (p *defaultPacker) writeSeqToWriter(writer *buffer.Writer, seq int32) {
	switch p.opts.seqBytes {
	case 1:
		writer.WriteInt8s(int8(seq))
	case 2:
		writer.WriteInt16s(p.opts.byteOrder, int16(seq))
	case 4:
		writer.WriteInt32s(p.opts.byteOrder, seq)
	}
}

// writeRouteToBuf 写路由到 bytes.Buffer
func (p *defaultPacker) writeRouteToBuf(buf *bytes.Buffer, route int32) error {
	switch p.opts.routeBytes {
	case 1:
		return binary.Write(buf, p.opts.byteOrder, int8(route))
	case 2:
		return binary.Write(buf, p.opts.byteOrder, int16(route))
	case 4:
		return binary.Write(buf, p.opts.byteOrder, route)
	}
	return nil
}

// writeSeqToBuf 写序列号到 bytes.Buffer
func (p *defaultPacker) writeSeqToBuf(buf *bytes.Buffer, seq int32) error {
	switch p.opts.seqBytes {
	case 1:
		return binary.Write(buf, p.opts.byteOrder, int8(seq))
	case 2:
		return binary.Write(buf, p.opts.byteOrder, int16(seq))
	case 4:
		return binary.Write(buf, p.opts.byteOrder, seq)
	}
	return nil
}

// readRoute 从 reader 读取路由
func (p *defaultPacker) readRoute(reader *bytes.Reader) (int32, error) {
	switch p.opts.routeBytes {
	case 1:
		var route int8
		if err := binary.Read(reader, p.opts.byteOrder, &route); err != nil {
			return 0, err
		}
		return int32(route), nil
	case 2:
		var route int16
		if err := binary.Read(reader, p.opts.byteOrder, &route); err != nil {
			return 0, err
		}
		return int32(route), nil
	case 4:
		var route int32
		if err := binary.Read(reader, p.opts.byteOrder, &route); err != nil {
			return 0, err
		}
		return route, nil
	}
	return 0, nil
}

// readSeq 从 reader 读取序列号
func (p *defaultPacker) readSeq(reader *bytes.Reader) (int32, error) {
	switch p.opts.seqBytes {
	case 0:
		return 0, nil
	case 1:
		var seq int8
		if err := binary.Read(reader, p.opts.byteOrder, &seq); err != nil {
			return 0, err
		}
		return int32(seq), nil
	case 2:
		var seq int16
		if err := binary.Read(reader, p.opts.byteOrder, &seq); err != nil {
			return 0, err
		}
		return int32(seq), nil
	case 4:
		var seq int32
		if err := binary.Read(reader, p.opts.byteOrder, &seq); err != nil {
			return 0, err
		}
		return seq, nil
	}
	return 0, nil
}

// getBuffer 从池中获取 buffer
func (p *defaultPacker) getBuffer() *bytes.Buffer {
	return p.bufferPool.Get().(*bytes.Buffer)
}

// putBuffer 将 buffer 放回池中
func (p *defaultPacker) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	p.bufferPool.Put(buf)
}

// makeHeartbeat 构建心跳包
func makeHeartbeat(byteOrder binary.ByteOrder) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Grow(defaultSizeBytes + defaultHeaderBytes)

	_ = binary.Write(buf, byteOrder, uint32(defaultHeaderBytes))
	_ = binary.Write(buf, byteOrder, uint8(heartbeatBit))

	return buf.Bytes()
}
