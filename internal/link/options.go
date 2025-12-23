package link

import (
	"github.com/dawnsgo/dawn/cluster"
	"github.com/dawnsgo/dawn/crypto"
	"github.com/dawnsgo/dawn/encoding"
	"github.com/dawnsgo/dawn/locate"
	"github.com/dawnsgo/dawn/registry"
)

type Options struct {
	InsID     string            // 实例ID
	InsKind   cluster.Kind      // 实例类型
	Codec     encoding.Codec    // 编解码器
	Locator   locate.Locator    // 定位器
	Registry  registry.Registry // 注册器
	Encryptor crypto.Encryptor  // 加密器
	Dispatch  cluster.Dispatch  // 无状态路由消息分发策略
}
