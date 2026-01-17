package errors

import (
	"fmt"
	"io"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/core/stack"
)

// ============================================================================
// 预定义错误（带错误码）
// ============================================================================

// 通用错误
var (
	ErrNil              = NewWithCode(codes.Unknown, "nil")
	ErrInvalidArgument  = NewWithCode(codes.InvalidArgument, "invalid argument")
	ErrInvalidPointer   = NewWithCode(codes.InvalidPointer, "invalid pointer")
	ErrInvalidFormat    = NewWithCode(codes.InvalidFormat, "invalid format")
	ErrIllegalRequest   = NewWithCode(codes.IllegalRequest, "illegal request")
	ErrIllegalOperation = NewWithCode(codes.IllegalInvoke, "illegal operation")
	ErrDeadlineExceeded = NewWithCode(codes.DeadlineExceeded, "deadline exceeded")
	ErrUnknownError     = NewWithCode(codes.Unknown, "unknown error")
	ErrNoOperationPermission = NewWithCode(codes.PermissionDenied, "no operation permission")
)

// 网络/连接错误
var (
	ErrInvalidReader       = NewWithCode(codes.InvalidReader, "invalid reader")
	ErrConnectionOpened    = NewWithCode(codes.ConnectionOpened, "connection is opened")
	ErrConnectionHanged    = NewWithCode(codes.ConnectionHanged, "connection is hanged")
	ErrConnectionClosed    = NewWithCode(codes.ConnectionClosed, "connection is closed")
	ErrConnectionNotOpened = NewWithCode(codes.ConnectionNotOpened, "connection is not opened")
	ErrConnectionNotHanged = NewWithCode(codes.ConnectionNotHanged, "connection is not hanged")
	ErrTooManyConnection   = NewWithCode(codes.TooManyConnections, "too many connection")
	ErrUnexpectedEOF       = NewWithCode(codes.UnexpectedEOF, "unexpected EOF")
)

// 会话/用户错误
var (
	ErrInvalidGID           = NewWithCode(codes.InvalidGateID, "invalid gate id")
	ErrInvalidNID           = NewWithCode(codes.InvalidNodeID, "invalid node id")
	ErrNotFoundSession      = NewWithCode(codes.SessionNotFound, "not found session")
	ErrInvalidSessionKind   = NewWithCode(codes.InvalidSessionKind, "invalid session kind")
	ErrNotFoundUserLocation = NewWithCode(codes.UserLocationNotFound, "not found user's location")
	ErrActorExists          = NewWithCode(codes.ActorExists, "actor exists")
	ErrNotFoundActor        = NewWithCode(codes.ActorNotFound, "not found actor")
	ErrNotBindActor         = NewWithCode(codes.ActorNotBound, "not bind actor")
)

// 路由/消息错误
var (
	ErrInvalidMessage     = NewWithCode(codes.InvalidMessage, "invalid message")
	ErrReceiveTargetEmpty = NewWithCode(codes.TargetEmpty, "the receive target is empty")
	ErrNotFoundRoute      = NewWithCode(codes.RouteNotFound, "not found route")
	ErrNotFoundEvent      = NewWithCode(codes.EventNotFound, "not found event")
	ErrSeqOverflow        = NewWithCode(codes.SeqOverflow, "seq overflow")
	ErrRouteOverflow      = NewWithCode(codes.RouteOverflow, "route overflow")
	ErrMessageTooLarge    = NewWithCode(codes.MessageTooLarge, "message too large")
	ErrInvalidDecoder     = NewWithCode(codes.InvalidDecoder, "invalid decoder")
	ErrInvalidScanner     = NewWithCode(codes.InvalidScanner, "invalid scanner")
	ErrUnregisterRoute    = NewWithCode(codes.RouteUnregistered, "unregistered route")
)

// 配置/初始化错误
var (
	ErrInvalidConfigContent    = NewWithCode(codes.InvalidConfig, "invalid config content")
	ErrNotFoundConfigSource    = NewWithCode(codes.ConfigSourceNotFound, "not found config source")
	ErrMissingTransporter      = NewWithCode(codes.MissingTransporter, "missing transporter")
	ErrMissingDiscovery        = NewWithCode(codes.MissingDiscovery, "missing discovery")
	ErrNotFoundLocator         = NewWithCode(codes.MissingLocator, "not found locator")
	ErrMissingResolver         = NewWithCode(codes.MissingResolver, "missing resolver")
	ErrMissingDispatchStrategy = NewWithCode(codes.MissingDispatchStrategy, "missing dispatch strategy")
	ErrMissingCacheInstance    = NewWithCode(codes.MissingCacheInstance, "missing cache instance")
	ErrMissingEventbusInstance = NewWithCode(codes.MissingEventbusInstance, "missing eventbus instance")
	ErrClientShut              = NewWithCode(codes.ClientShut, "client is shut")
	ErrClientClosed            = NewWithCode(codes.ClientShut, "client is closed")
	ErrServerClosed            = NewWithCode(codes.ServerClosed, "server is closed")
	ErrSyncerClosed            = NewWithCode(codes.SyncerClosed, "syncer is closed")
)

// 服务/集群错误
var (
	ErrNotFoundEndpoint        = NewWithCode(codes.EndpointNotFound, "not found endpoint")
	ErrNotFoundServiceAddress  = NewWithCode(codes.ServiceAddressNotFound, "not found service address")
	ErrServiceRegisterFailed   = NewWithCode(codes.ServiceRegisterFailed, "service register failed")
	ErrServiceDeregisterFailed = NewWithCode(codes.ServiceDeregisterFailed, "service deregister failed")
	ErrInvalidServiceDesc      = NewWithCode(codes.InvalidServiceDesc, "invalid service desc")
	ErrNotFoundIPAddress       = NewWithCode(codes.ServiceAddressNotFound, "not found ip address")
)

// 安全/加密错误
var (
	ErrInvalidPublicKey  = NewWithCode(codes.InvalidPublicKey, "invalid public key")
	ErrInvalidPrivateKey = NewWithCode(codes.InvalidPrivateKey, "invalid private key")
	ErrInvalidSignature  = NewWithCode(codes.InvalidSignature, "invalid signature")
	ErrInvalidCertFile   = NewWithCode(codes.InvalidCertFile, "invalid cert file")
	ErrInvalidKeyFormat  = NewWithCode(codes.InvalidKeyFormat, "invalid key format")
	ErrBlockSizeTooLarge = NewWithCode(codes.BlockSizeTooLarge, "block size too large for RSA key")
)

// 编码/序列化错误
var (
	ErrCodecNotRegistered     = NewWithCode(codes.CodecNotRegistered, "codec not registered")
	ErrSignerNotRegistered    = NewWithCode(codes.SignerNotRegistered, "signer not registered")
	ErrEncryptorNotRegistered = NewWithCode(codes.EncryptorNotRegistered, "encryptor not registered")
	ErrProtoMarshalError      = NewWithCode(codes.ProtoMarshalError, "proto marshal error")
	ErrProtoUnmarshalError    = NewWithCode(codes.ProtoUnmarshalError, "proto unmarshal error")
)

// ============================================================================
// 错误创建函数
// ============================================================================

// NewSimple 新建一个简单错误
func NewSimple(text string) *Error {
	return &Error{text: text}
}

// NewWithCode 新建一个带错误码的错误
func NewWithCode(code *codes.Code, text string) *Error {
	return &Error{code: code, text: text}
}

// NewError 新建一个错误
// 可传入以下参数：
// text : 文本字符串
// code : 错误码
// error: 原生错误
func NewError(args ...any) *Error {
	e := &Error{}

	for _, arg := range args {
		switch v := arg.(type) {
		case error:
			e.err = v
		case string:
			e.text = v
		case *codes.Code:
			e.code = v
		}
	}

	return e
}

// NewErrorWithStack 新建一个带堆栈的错误
// 可传入以下参数：
// text : 文本字符串
// code : 错误码
// error: 原生错误
func NewErrorWithStack(args ...any) *Error {
	e := &Error{stack: stack.Callers(1, stack.Full)}

	for _, arg := range args {
		switch v := arg.(type) {
		case error:
			e.err = v
		case string:
			e.text = v
		case *codes.Code:
			e.code = v
		}
	}

	return e
}

// Wrap 包装一个错误，添加上下文信息
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{err: err, text: message}
}

// WrapWithCode 包装一个错误，添加错误码和上下文信息
func WrapWithCode(err error, code *codes.Code, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{err: err, code: code, text: message}
}

// ============================================================================
// 错误辅助函数
// ============================================================================

// Code 返回错误码
func Code(err error) *codes.Code {
	if err != nil {
		if e, ok := err.(interface{ Code() *codes.Code }); ok {
			return e.Code()
		}
	}

	return nil
}

// Next 返回下一个错误
func Next(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(interface{ Next() error }); ok {
		return e.Next()
	}

	return nil
}

// Cause 返回根因错误
func Cause(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(interface{ Cause() error }); ok {
		return e.Cause()
	}

	return err
}

// UnwrapError 解包错误（返回下一层错误）
func UnwrapError(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(interface{ Unwrap() error }); ok {
		return e.Unwrap()
	}

	return nil
}

// Stack 返回堆栈
func Stack(err error) *stack.Stack {
	if err == nil {
		return nil
	}

	if e, ok := err.(interface{ Stack() *stack.Stack }); ok {
		return e.Stack()
	}

	return nil
}

// Replace 替换文本
func Replace(err error, text string, condition ...codes.Code) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(interface {
		Replace(text string, condition ...codes.Code) error
	}); ok {
		return e.Replace(text, condition...)
	}

	return err
}

// IsErr 判断错误是否是指定错误（支持链式错误和错误码比较）
func IsErr(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}

	// 先使用标准库 Is 检查
	if Is(err, target) {
		return true
	}

	// 比较错误码
	errCode := Code(err)
	targetCode := Code(target)
	if errCode != nil && targetCode != nil && errCode.Code() == targetCode.Code() {
		return true
	}

	// 递归检查链式错误
	if next := Next(err); next != nil {
		return IsErr(next, target)
	}

	return false
}

// IsCode 判断错误是否是指定错误码
func IsCode(err error, code *codes.Code) bool {
	return codes.IsCode(err, code)
}

// IsNetworkError 判断是否是网络错误
func IsNetworkError(err error) bool {
	return codes.IsNetworkErr(err)
}

// IsSessionError 判断是否是会话错误
func IsSessionError(err error) bool {
	return codes.IsSessionErr(err)
}

// IsRouteError 判断是否是路由错误
func IsRouteError(err error) bool {
	return codes.IsRouteErr(err)
}

// IsConfigError 判断是否是配置错误
func IsConfigError(err error) bool {
	return codes.IsConfigErr(err)
}

// IsServiceError 判断是否是服务错误
func IsServiceError(err error) bool {
	return codes.IsServiceErr(err)
}

// IsSecurityError 判断是否是安全错误
func IsSecurityError(err error) bool {
	return codes.IsSecurityErr(err)
}

// IsEncodingError 判断是否是编码错误
func IsEncodingError(err error) bool {
	return codes.IsEncodingErr(err)
}

// ============================================================================
// Error 结构体
// ============================================================================

type Error struct {
	err   error
	text  string
	code  *codes.Code
	stack *stack.Stack
}

func (e *Error) Error() (text string) {
	if e == nil {
		return
	}

	if e.code != nil && e.code != codes.OK {
		text = e.code.String()
	}

	if e.text != "" {
		if text != "" {
			text += ": "
		}
		text += e.text
	}

	if e.err != nil && e.err.Error() != "" {
		if text != "" {
			text += ": "
		}
		text += e.err.Error()
	}

	return
}

// Code 返回错误码
func (e *Error) Code() *codes.Code {
	if e == nil {
		return nil
	}

	return e.code
}

// Next 返回下一个错误
func (e *Error) Next() error {
	if e == nil {
		return nil
	}

	return e.err
}

// Cause 返回根因错误
func (e *Error) Cause() error {
	if e == nil {
		return nil
	}

	if e.err == nil {
		return e
	}

	cause := e.err
	for cause != nil {
		if ce, ok := cause.(interface{ Cause() error }); ok {
			cause = ce.Cause()
		} else {
			break
		}
	}

	return cause
}

// Stack 返回堆栈
func (e *Error) Stack() *stack.Stack {
	if e == nil {
		return nil
	}

	return e.stack
}

// Unwrap 解包错误
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.err
}

// Replace 替换文本
func (e *Error) Replace(text string, condition ...*codes.Code) error {
	if e == nil {
		return nil
	}

	if len(condition) == 0 || condition[0] == e.code {
		e.text = text
	}

	return e
}

// String 格式化错误信息
func (e *Error) String() string {
	return fmt.Sprintf("%+v", e)
}

func (e *Error) error() (text string) {
	if e == nil {
		return
	}

	text = e.text
	if text == "" && e.code != codes.OK {
		text = e.code.String()
	}

	return
}

// Format 格式化输出
// %s : 打印本级错误信息
// %v : 打印所有错误信息
// %+v: 打印所有错误信息和堆栈信息
func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil {
		return
	}

	switch verb {
	case 'v':
		if s.Flag('+') {
			var (
				i    int
				next error = e
			)

			io.WriteString(s, e.Error()+"\nStack:\n")
			for next != nil {
				i++
				if n, ok := next.(*Error); ok {
					fmt.Fprintf(s, "%d. %s\n", i, n.error())
					for i, f := range n.stack.Frames() {
						fmt.Fprintf(s, "\t%d). %s\n\t%s:%d\n",
							i+1,
							f.Function,
							f.File,
							f.Line,
						)
					}
					next = n.Next()
				} else {
					fmt.Fprintf(s, "%d. %s\n", i, next.Error())
					break
				}
			}
		} else {
			io.WriteString(s, e.Error())
		}
	case 's':
		if e.text != "" {
			io.WriteString(s, e.text)
		} else {
			e.code.Format(s, verb)
		}
	}
}
