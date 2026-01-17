package crypto

import (
	"fmt"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
)

type Signer interface {
	// Name 名称
	Name() string
	// Sign 签名
	Sign(data []byte) ([]byte, error)
	// Verify 验签
	Verify(data []byte, signature []byte) (bool, error)
}

var signers = make(map[string]Signer)

// RegisterSigner 注册签名器
func RegisterSigner(signer Signer) {
	if signer == nil {
		panic("can't register a nil signer")
	}

	name := signer.Name()

	if name == "" {
		panic("can't register a signer without name")
	}

	if _, ok := signers[name]; ok {
		log.Warnf("the old %s signer will be overwritten", name)
	}

	signers[name] = signer
}

// InvokeSigner 调用签名器
func InvokeSigner(name string) Signer {
	signer, ok := signers[name]
	if !ok {
		panic(fmt.Sprintf("%s signer is not registered", name))
	}

	return signer
}

// InvokeSignerWithError 调用签名器（返回错误）
func InvokeSignerWithError(name string) (Signer, error) {
	signer, ok := signers[name]
	if !ok {
		return nil, errors.NewWithCode(codes.SignerNotRegistered, fmt.Sprintf("%s signer is not registered", name))
	}

	return signer, nil
}
