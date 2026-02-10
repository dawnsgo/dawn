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

// RouteCodec 路由编解码器接口，消除重复的 switch 分支
type RouteCodec interface {
	// EncodeToWriter 将路由编码写入 buffer.Writer
	EncodeToWriter(writer *buffer.Writer, route int32)
	// EncodeToBuf 将路由编码写入 bytes.Buffer
	EncodeToBuf(buf *bytes.Buffer, route int32) error
	// Decode 从 reader 中解码路由
	Decode(reader *bytes.Reader) (int32, error)
	// Size 返回路由占用的字节数
	Size() int
}

// SeqCodec 序列号编解码器接口
type SeqCodec interface {
	// EncodeToWriter 将序列号编码写入 buffer.Writer
	EncodeToWriter(writer *buffer.Writer, seq int32)
	// EncodeToBuf 将序列号编码写入 bytes.Buffer
	EncodeToBuf(buf *bytes.Buffer, seq int32) error
	// Decode 从 reader 中解码序列号
	Decode(reader *bytes.Reader) (int32, error)
	// Size 返回序列号占用的字节数
	Size() int
}

type defaultPacker struct {
	opts       *options
	heartbeat  []byte
	routeCodec RouteCodec
	seqCodec   SeqCodec
	bufferPool sync.Pool // 复用 bytes.Buffer
}

// NewPacker 创建一个新的消息打包器
func NewPacker(opts ...Option) (Packer, error) {
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

	p := &defaultPacker{
		opts:       o,
		heartbeat:  makeHeartbeat(o.byteOrder),
		routeCodec: newIntCodec(o.routeBytes, o.byteOrder),
		seqCodec:   newIntCodec(o.seqBytes, o.byteOrder),
		bufferPool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, defaultSizeBytes+defaultHeaderBytes+o.routeBytes+o.seqBytes+256))
			},
		},
	}

	return p, nil
}

// MustNewPacker 创建一个新的消息打包器，失败时 panic
// 这是一个便捷函数，适用于初始化阶段，错误表示程序配置有误
func MustNewPacker(opts ...Option) Packer {
	p, err := NewPacker(opts...)
	if err != nil {
		panic("dawn/packet: create packer failed: " + err.Error())
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

// 使用 codec 委托，消除重复的 switch 分支
func (p *defaultPacker) writeRouteToWriter(writer *buffer.Writer, route int32) {
	p.routeCodec.EncodeToWriter(writer, route)
}

func (p *defaultPacker) writeSeqToWriter(writer *buffer.Writer, seq int32) {
	p.seqCodec.EncodeToWriter(writer, seq)
}

func (p *defaultPacker) writeRouteToBuf(buf *bytes.Buffer, route int32) error {
	return p.routeCodec.EncodeToBuf(buf, route)
}

func (p *defaultPacker) writeSeqToBuf(buf *bytes.Buffer, seq int32) error {
	return p.seqCodec.EncodeToBuf(buf, seq)
}

func (p *defaultPacker) readRoute(reader *bytes.Reader) (int32, error) {
	return p.routeCodec.Decode(reader)
}

func (p *defaultPacker) readSeq(reader *bytes.Reader) (int32, error) {
	return p.seqCodec.Decode(reader)
}

// ==================== IntCodec 通用整数编解码器 ====================
// 统一处理 1/2/4 字节整数的编解码，消除重复的 switch 逻辑

type intCodec struct {
	size      int
	byteOrder binary.ByteOrder
}

// newIntCodec 创建整数编解码器
func newIntCodec(size int, byteOrder binary.ByteOrder) *intCodec {
	return &intCodec{size: size, byteOrder: byteOrder}
}

func (c *intCodec) Size() int {
	return c.size
}

func (c *intCodec) EncodeToWriter(writer *buffer.Writer, val int32) {
	switch c.size {
	case 1:
		writer.WriteInt8s(int8(val))
	case 2:
		writer.WriteInt16s(c.byteOrder, int16(val))
	case 4:
		writer.WriteInt32s(c.byteOrder, val)
	}
}

func (c *intCodec) EncodeToBuf(buf *bytes.Buffer, val int32) error {
	switch c.size {
	case 0:
		return nil
	case 1:
		return binary.Write(buf, c.byteOrder, int8(val))
	case 2:
		return binary.Write(buf, c.byteOrder, int16(val))
	case 4:
		return binary.Write(buf, c.byteOrder, val)
	}
	return nil
}

func (c *intCodec) Decode(reader *bytes.Reader) (int32, error) {
	switch c.size {
	case 0:
		return 0, nil
	case 1:
		var v int8
		if err := binary.Read(reader, c.byteOrder, &v); err != nil {
			return 0, err
		}
		return int32(v), nil
	case 2:
		var v int16
		if err := binary.Read(reader, c.byteOrder, &v); err != nil {
			return 0, err
		}
		return int32(v), nil
	case 4:
		var v int32
		if err := binary.Read(reader, c.byteOrder, &v); err != nil {
			return 0, err
		}
		return v, nil
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
