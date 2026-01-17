package codes

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ============================================================================
// 错误码分类说明
// ============================================================================
// 0:       OK - 成功
// 1-99:    通用错误 - 取消、未知、超时、参数无效等
// 100-199: 网络/连接错误 - 连接关闭、太多连接等
// 200-299: 会话/用户错误 - 会话不存在、用户未找到等
// 300-399: 路由/消息错误 - 路由不存在、消息格式错误等
// 400-499: 配置/初始化错误 - 配置无效、组件缺失等
// 500-599: 服务/集群错误 - 服务注册失败、节点不存在等
// 600-699: 安全/加密错误 - 签名无效、证书错误等
// 700-799: 编码/序列化错误 - 编解码器未注册等
// ============================================================================

// 错误码分类范围
const (
	CategoryGeneral    = 0    // 通用错误起始
	CategoryNetwork    = 100  // 网络错误起始
	CategorySession    = 200  // 会话错误起始
	CategoryRoute      = 300  // 路由错误起始
	CategoryConfig     = 400  // 配置错误起始
	CategoryService    = 500  // 服务错误起始
	CategorySecurity   = 600  // 安全错误起始
	CategoryEncoding   = 700  // 编码错误起始
	CategoryCustomBase = 1000 // 用户自定义错误码起始
)

// ============================================================================
// 通用错误码 (0-99)
// ============================================================================

var (
	OK                = NewCode(0, "ok")
	Canceled          = NewCode(1, "canceled")
	Unknown           = NewCode(2, "unknown")
	InvalidArgument   = NewCode(3, "invalid argument")
	DeadlineExceeded  = NewCode(4, "deadline exceeded")
	NotFound          = NewCode(5, "not found")
	InternalError     = NewCode(6, "internal error")
	Unauthorized      = NewCode(7, "unauthorized")
	IllegalInvoke     = NewCode(8, "illegal invoke")
	IllegalRequest    = NewCode(9, "illegal request")
	TooManyRequests   = NewCode(10, "too many requests")
	InvalidPointer    = NewCode(11, "invalid pointer")
	InvalidFormat     = NewCode(12, "invalid format")
	PermissionDenied  = NewCode(13, "permission denied")
	ResourceExhausted = NewCode(14, "resource exhausted")
	Unavailable       = NewCode(15, "unavailable")
	Unimplemented     = NewCode(16, "unimplemented")
	InvalidSeekWhence = NewCode(17, "invalid seek whence")
	NegativePosition  = NewCode(18, "negative position")
	IndexOverflow     = NewCode(19, "index overflow")
	BadSyntax         = NewCode(20, "bad syntax")
)

// ============================================================================
// 网络/连接错误码 (100-199)
// ============================================================================

var (
	ConnectionOpened    = NewCode(100, "connection is opened")
	ConnectionHanged    = NewCode(101, "connection is hanged")
	ConnectionClosed    = NewCode(102, "connection is closed")
	ConnectionNotOpened = NewCode(103, "connection is not opened")
	ConnectionNotHanged = NewCode(104, "connection is not hanged")
	TooManyConnections  = NewCode(105, "too many connections")
	NetworkError        = NewCode(106, "network error")
	ReadError           = NewCode(107, "read error")
	WriteError          = NewCode(108, "write error")
	InvalidReader       = NewCode(109, "invalid reader")
	UnexpectedEOF       = NewCode(110, "unexpected EOF")
)

// ============================================================================
// 会话/用户错误码 (200-299)
// ============================================================================

var (
	SessionNotFound      = NewCode(200, "session not found")
	InvalidSessionKind   = NewCode(201, "invalid session kind")
	UserNotFound         = NewCode(202, "user not found")
	UserLocationNotFound = NewCode(203, "user location not found")
	InvalidGateID        = NewCode(204, "invalid gate id")
	InvalidNodeID        = NewCode(205, "invalid node id")
	ActorExists          = NewCode(206, "actor exists")
	ActorNotFound        = NewCode(207, "actor not found")
	ActorNotBound        = NewCode(208, "actor not bound")
)

// ============================================================================
// 路由/消息错误码 (300-399)
// ============================================================================

var (
	RouteNotFound    = NewCode(300, "route not found")
	RouteUnregistered = NewCode(301, "route unregistered")
	RouteOverflow    = NewCode(302, "route overflow")
	EventNotFound    = NewCode(303, "event not found")
	InvalidMessage   = NewCode(304, "invalid message")
	MessageTooLarge  = NewCode(305, "message too large")
	SeqOverflow      = NewCode(306, "seq overflow")
	TargetEmpty      = NewCode(307, "receive target is empty")
	InvalidDecoder   = NewCode(308, "invalid decoder")
	InvalidScanner   = NewCode(309, "invalid scanner")
)

// ============================================================================
// 配置/初始化错误码 (400-499)
// ============================================================================

var (
	InvalidConfig           = NewCode(400, "invalid config")
	ConfigSourceNotFound    = NewCode(401, "config source not found")
	MissingComponent        = NewCode(402, "missing component")
	MissingTransporter      = NewCode(403, "missing transporter")
	MissingDiscovery        = NewCode(404, "missing discovery")
	MissingLocator          = NewCode(405, "missing locator")
	MissingResolver         = NewCode(406, "missing resolver")
	MissingDispatchStrategy = NewCode(407, "missing dispatch strategy")
	MissingCacheInstance    = NewCode(408, "missing cache instance")
	MissingEventbusInstance = NewCode(409, "missing eventbus instance")
	ComponentClosed         = NewCode(410, "component closed")
	ClientShut              = NewCode(411, "client is shut")
	ServerClosed            = NewCode(412, "server is closed")
	SyncerClosed            = NewCode(413, "syncer is closed")
	InvalidFilePath         = NewCode(414, "invalid file path")
	InvalidServerURL        = NewCode(415, "invalid server url")
)

// ============================================================================
// 服务/集群错误码 (500-599)
// ============================================================================

var (
	EndpointNotFound        = NewCode(500, "endpoint not found")
	ServiceAddressNotFound  = NewCode(501, "service address not found")
	ServiceRegisterFailed   = NewCode(502, "service register failed")
	ServiceDeregisterFailed = NewCode(503, "service deregister failed")
	InvalidServiceDesc      = NewCode(504, "invalid service desc")
)

// ============================================================================
// 安全/加密错误码 (600-699)
// ============================================================================

var (
	InvalidPublicKey   = NewCode(600, "invalid public key")
	InvalidPrivateKey  = NewCode(601, "invalid private key")
	InvalidSignature   = NewCode(602, "invalid signature")
	InvalidCertFile    = NewCode(603, "invalid cert file")
	EncryptionError    = NewCode(604, "encryption error")
	DecryptionError    = NewCode(605, "decryption error")
	InvalidKeyFormat   = NewCode(606, "invalid key format")
	BlockSizeTooLarge  = NewCode(607, "block size too large for RSA key")
)

// ============================================================================
// 编码/序列化错误码 (700-799)
// ============================================================================

var (
	CodecNotRegistered     = NewCode(700, "codec not registered")
	SignerNotRegistered    = NewCode(701, "signer not registered")
	EncryptorNotRegistered = NewCode(702, "encryptor not registered")
	ProtoMarshalError      = NewCode(703, "proto marshal error")
	ProtoUnmarshalError    = NewCode(704, "proto unmarshal error")
)

// ============================================================================
// Code 结构体和方法
// ============================================================================

type Code struct {
	code    int
	message string
}

// NewCode 新建一个错误码
func NewCode(code int, message ...string) *Code {
	if len(message) > 0 {
		return &Code{code: code, message: message[0]}
	} else {
		return &Code{code: code}
	}
}

// Code 返回错误码
func (c *Code) Code() int {
	return c.code
}

// WithCode 替换新的错误码
func (c *Code) WithCode(code int) *Code {
	return &Code{
		code:    code,
		message: c.message,
	}
}

// Message 返回错误码消息
func (c *Code) Message() string {
	return c.message
}

// WithMessage 替换新的错误码消息
func (c *Code) WithMessage(message string) *Code {
	return &Code{
		code:    c.code,
		message: message,
	}
}

// WithMessagef 格式化替换新的错误码消息
func (c *Code) WithMessagef(format string, a ...any) *Code {
	return c.WithMessage(fmt.Sprintf(format, a...))
}

// String 格式化错误码
func (c *Code) String() string {
	return fmt.Sprintf("code error: code = %d desc = %s", c.code, c.message)
}

// Format 格式化输出
// %s : 打印错误码和错误消息
// %v : 打印错误码、错误消息、错误详情
func (c *Code) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		if c.message != "" {
			io.WriteString(s, fmt.Sprintf("%d:%s", c.code, c.message))
		} else {
			io.WriteString(s, fmt.Sprintf("%d", c.code))
		}
	case 'v':
		io.WriteString(s, c.String())
	}
}

// Err 转错误消息
func (c *Code) Err() error {
	if c.code == OK.Code() {
		return nil
	}

	return &Error{code: c}
}

// Category 返回错误码所属分类
func (c *Code) Category() int {
	if c.code < CategoryNetwork {
		return CategoryGeneral
	} else if c.code < CategorySession {
		return CategoryNetwork
	} else if c.code < CategoryRoute {
		return CategorySession
	} else if c.code < CategoryConfig {
		return CategoryRoute
	} else if c.code < CategoryService {
		return CategoryConfig
	} else if c.code < CategorySecurity {
		return CategoryService
	} else if c.code < CategoryEncoding {
		return CategorySecurity
	} else if c.code < CategoryCustomBase {
		return CategoryEncoding
	}
	return CategoryCustomBase
}

// IsGeneral 是否是通用错误
func (c *Code) IsGeneral() bool {
	return c.Category() == CategoryGeneral
}

// IsNetworkError 是否是网络错误
func (c *Code) IsNetworkError() bool {
	return c.Category() == CategoryNetwork
}

// IsSessionError 是否是会话错误
func (c *Code) IsSessionError() bool {
	return c.Category() == CategorySession
}

// IsRouteError 是否是路由错误
func (c *Code) IsRouteError() bool {
	return c.Category() == CategoryRoute
}

// IsConfigError 是否是配置错误
func (c *Code) IsConfigError() bool {
	return c.Category() == CategoryConfig
}

// IsServiceError 是否是服务错误
func (c *Code) IsServiceError() bool {
	return c.Category() == CategoryService
}

// IsSecurityError 是否是安全错误
func (c *Code) IsSecurityError() bool {
	return c.Category() == CategorySecurity
}

// IsEncodingError 是否是编码错误
func (c *Code) IsEncodingError() bool {
	return c.Category() == CategoryEncoding
}

// ============================================================================
// Error 结构体
// ============================================================================

type Error struct {
	code *Code
}

// Error error interface implementation
func (e *Error) Error() string {
	return e.code.String()
}

// Code 返回错误码
func (e *Error) Code() *Code {
	return e.code
}

// ============================================================================
// 辅助函数
// ============================================================================

// Convert 将错误信息转换为错误码
func Convert(err error) *Code {
	if err == nil {
		return OK
	}

	if e, ok := err.(interface{ Code() *Code }); ok {
		return e.Code()
	}

	text := err.Error()
	flag := "code error:"
	index := strings.Index(text, flag)

	if index == -1 {
		return Unknown
	}

	after, found := strings.CutPrefix(text[index+len(flag):], " code = ")
	if !found {
		return Unknown
	}

	elements := strings.SplitN(after, " ", 2)
	if len(elements) != 2 {
		return Unknown
	}

	code, err := strconv.Atoi(elements[0])
	if err != nil {
		return Unknown
	}

	after, found = strings.CutPrefix(elements[1], "desc = ")
	if !found {
		return Unknown
	}

	return NewCode(code, after)
}

// IsCode 判断错误是否是指定错误码
func IsCode(err error, code *Code) bool {
	if err == nil {
		return code == OK || code.code == 0
	}
	c := Convert(err)
	return c.code == code.code
}

// CategoryOf 获取错误的分类
func CategoryOf(err error) int {
	if err == nil {
		return CategoryGeneral
	}
	return Convert(err).Category()
}

// IsNetworkErr 判断错误是否是网络错误
func IsNetworkErr(err error) bool {
	return CategoryOf(err) == CategoryNetwork
}

// IsSessionErr 判断错误是否是会话错误
func IsSessionErr(err error) bool {
	return CategoryOf(err) == CategorySession
}

// IsRouteErr 判断错误是否是路由错误
func IsRouteErr(err error) bool {
	return CategoryOf(err) == CategoryRoute
}

// IsConfigErr 判断错误是否是配置错误
func IsConfigErr(err error) bool {
	return CategoryOf(err) == CategoryConfig
}

// IsServiceErr 判断错误是否是服务错误
func IsServiceErr(err error) bool {
	return CategoryOf(err) == CategoryService
}

// IsSecurityErr 判断错误是否是安全错误
func IsSecurityErr(err error) bool {
	return CategoryOf(err) == CategorySecurity
}

// IsEncodingErr 判断错误是否是编码错误
func IsEncodingErr(err error) bool {
	return CategoryOf(err) == CategoryEncoding
}
