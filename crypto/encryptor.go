package crypto

import (
	"fmt"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
)

type Encryptor interface {
	// Name 名称
	Name() string
	// Encrypt 加密
	Encrypt(data []byte) ([]byte, error)
	// Decrypt 解密
	Decrypt(data []byte) ([]byte, error)
}

var encryptors = make(map[string]Encryptor)

// RegisterEncryptor 注册加密器
func RegisterEncryptor(encryptor Encryptor) {
	if encryptor == nil {
		panic("can't register a nil encryptor")
	}

	name := encryptor.Name()

	if name == "" {
		panic("can't register an encryptor without name")
	}

	if _, ok := encryptors[name]; ok {
		log.Warnf("the old %s encryptor will be overwritten", name)
	}

	encryptors[name] = encryptor
}

// InvokeEncryptor 调用加密器
func InvokeEncryptor(name string) Encryptor {
	encryptor, ok := encryptors[name]
	if !ok {
		panic(fmt.Sprintf("%s encryptor is not registered", name))
	}

	return encryptor
}

// InvokeEncryptorWithError 调用加密器（返回错误）
func InvokeEncryptorWithError(name string) (Encryptor, error) {
	encryptor, ok := encryptors[name]
	if !ok {
		return nil, errors.NewWithCode(codes.EncryptorNotRegistered, fmt.Sprintf("%s encryptor is not registered", name))
	}

	return encryptor, nil
}
