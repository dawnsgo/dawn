/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/5/14 10:47 上午
 * @Desc: TODO
 */

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

// Marshal 编码
func Marshal(v any) ([]byte, error) {
	return DefaultCodec.Marshal(v)
}

// Unmarshal 解码
func Unmarshal(data []byte, v any) error {
	return DefaultCodec.Unmarshal(data, v)
}
