package link

import (
	"github.com/dawnsgo/dawn/v2/cluster"
	"github.com/dawnsgo/dawn/v2/crypto"
	"github.com/dawnsgo/dawn/v2/encoding"
	"github.com/dawnsgo/dawn/v2/locate"
	"github.com/dawnsgo/dawn/v2/registry"
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
