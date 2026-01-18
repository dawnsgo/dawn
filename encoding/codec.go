// Package encoding 提供统一的编解码器接口和多种编码格式支持。
// 支持 JSON、Protobuf、TOML、XML、YAML、MsgPack 等格式。
//
// 使用示例：
//
//	// 注册自定义编解码器
//	encoding.Register(myCodec)
//
//	// 获取并使用编解码器
//	codec := encoding.Invoke("json")
//	data, err := codec.Marshal(obj)
package encoding

import (
	"fmt"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/encoding/json"
	"github.com/dawnsgo/dawn/encoding/msgpack"
	"github.com/dawnsgo/dawn/encoding/proto"
	"github.com/dawnsgo/dawn/encoding/toml"
	"github.com/dawnsgo/dawn/encoding/xml"
	"github.com/dawnsgo/dawn/encoding/yaml"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
)

var codecs = make(map[string]Codec)

func init() {
	Register(json.DefaultCodec)
	Register(proto.DefaultCodec)
	Register(toml.DefaultCodec)
	Register(xml.DefaultCodec)
	Register(yaml.DefaultCodec)
	Register(msgpack.DefaultCodec)
}

// Codec 编解码器接口，定义了数据序列化和反序列化的标准方法。
// 实现该接口的类型可以注册到编码包中，供框架统一使用。
type Codec interface {
	// Name 返回编解码器的名称，用于注册和查找
	Name() string
	// Marshal 将对象编码为字节数组
	Marshal(v any) ([]byte, error)
	// Unmarshal 将字节数组解码为对象
	Unmarshal(data []byte, v any) error
}

// Register 注册一个编解码器。
// 如果名称已存在，会覆盖原有的编解码器。
// 如果 codec 为 nil 或名称为空，会触发 panic。
func Register(codec Codec) {
	if codec == nil {
		panic("can't register a nil codec")
	}

	name := codec.Name()

	if name == "" {
		panic("can't register a codec without name")
	}

	if _, ok := codecs[name]; ok {
		log.Warnf("the old %s codec will be overwritten", name)
	}

	codecs[name] = codec
}

// Invoke 根据名称获取已注册的编解码器。
// 如果编解码器不存在，会触发 panic。
// 如需获取错误信息而非 panic，请使用 InvokeWithError。
func Invoke(name string) Codec {
	codec, ok := codecs[name]
	if !ok {
		panic(fmt.Sprintf("%s codec is not registered", name))
	}

	return codec
}

// InvokeWithError 根据名称获取已注册的编解码器，返回错误而非 panic。
// 如果编解码器不存在，返回 CodecNotRegistered 错误。
func InvokeWithError(name string) (Codec, error) {
	codec, ok := codecs[name]
	if !ok {
		return nil, errors.NewWithCode(codes.CodecNotRegistered, fmt.Sprintf("%s codec is not registered", name))
	}

	return codec, nil
}
