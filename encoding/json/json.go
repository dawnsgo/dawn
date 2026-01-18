/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/5/14 10:42 上午
 * @Desc: JSON编解码器实现，基于sonic库提供高性能的JSON序列化和反序列化功能
 */

package json

import (
	"github.com/bytedance/sonic"
)

const Name = "json"

var DefaultCodec = &codec{}

type codec struct{}

// Name 编解码器名称
func (codec) Name() string {
	return Name
}

// Marshal 编码
func (codec) Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal 解码
func (codec) Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// Marshal 编码
func Marshal(v any) ([]byte, error) {
	return DefaultCodec.Marshal(v)
}

// Unmarshal 解码
func Unmarshal(data []byte, v any) error {
	return DefaultCodec.Unmarshal(data, v)
}
