// Package proto 提供 Protocol Buffers 格式的编解码器实现。
// 支持对实现 proto.Message 接口的对象进行序列化和反序列化。
//
// 使用示例：
//
//	// 使用默认编解码器
//	data, err := proto.Marshal(message)
//	err = proto.Unmarshal(data, &message)
//
//	// 或通过 encoding 包使用
//	codec := encoding.Invoke("proto")
//	data, err := codec.Marshal(message)
package proto

import (
	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/errors"
	"google.golang.org/protobuf/proto"
)

const Name = "proto"

var DefaultCodec = &codec{}

type codec struct{}

// Name 编解码器名称
func (codec) Name() string {
	return Name
}

// Marshal 编码
func (codec) Marshal(v any) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, errors.NewWithCode(codes.ProtoMarshalError, "can't marshal a value that not implements proto.Message interface")
	}

	return proto.Marshal(msg)
}

// Unmarshal 解码
func (codec) Unmarshal(data []byte, v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.NewWithCode(codes.ProtoUnmarshalError, "can't unmarshal to a value that not implements proto.Message")
	}

	return proto.Unmarshal(data, msg)
}

// Marshal 使用默认编解码器将 proto.Message 编码为字节数组。
// 如果 v 不实现 proto.Message 接口，返回 ProtoMarshalError 错误。
func Marshal(v any) ([]byte, error) {
	return DefaultCodec.Marshal(v)
}

// Unmarshal 使用默认编解码器将字节数组解码为 proto.Message。
// 如果 v 不实现 proto.Message 接口，返回 ProtoUnmarshalError 错误。
func Unmarshal(data []byte, v any) error {
	return DefaultCodec.Unmarshal(data, v)
}
