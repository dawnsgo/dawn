package redis

import (
	"time"

	"github.com/dawnsgo/dawn/etc"
	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Redis 部署模式
// ============================================================================

// Mode Redis部署模式
type Mode string

const (
	// ModeSingle 单机模式
	ModeSingle Mode = "single"
	// ModeSentinel 哨兵模式
	ModeSentinel Mode = "sentinel"
	// ModeCluster 集群模式
	ModeCluster Mode = "cluster"
)

const (
	defaultAddr          = "127.0.0.1:6379"
	defaultDB            = 0
	defaultMaxRetries    = 3
	defaultPrefix        = "dawn:cache"
	defaultNilValue      = "cache@nil"
	defaultNilExpiration = "10s"
	defaultMinExpiration = "1h"
	defaultMaxExpiration = "24h"
)

const (
	defaultModeKey          = "etc.cache.redis.mode"
	defaultAddrsKey         = "etc.cache.redis.addrs"
	defaultDBKey            = "etc.cache.redis.db"
	defaultMaxRetriesKey    = "etc.cache.redis.maxRetries"
	defaultPrefixKey        = "etc.cache.redis.prefix"
	defaultUsernameKey      = "etc.cache.redis.username"
	defaultPasswordKey      = "etc.cache.redis.password"
	defaultCertFileKey      = "etc.cache.redis.certFile"
	defaultKeyFileKey       = "etc.cache.redis.keyFile"
	defaultCAFileKey        = "etc.cache.redis.caFile"
	defaultNilValueKey      = "etc.cache.redis.nilValue"
	defaultNilExpirationKey = "etc.cache.redis.nilExpiration"
	defaultMinExpirationKey = "etc.cache.redis.minExpiration"
	defaultMaxExpirationKey = "etc.cache.redis.maxExpiration"
	// 哨兵模式配置键
	defaultMasterNameKey = "etc.cache.redis.sentinel.masterName"
	defaultSentinelsKey  = "etc.cache.redis.sentinel.sentinels"
	// 集群模式配置键
	defaultClusterAddrsKey = "etc.cache.redis.cluster.addrs"
	// 连接池配置键
	defaultPoolSizeKey     = "etc.cache.redis.poolSize"
	defaultMinIdleConnsKey = "etc.cache.redis.minIdleConns"
	defaultConnMaxLifeKey  = "etc.cache.redis.connMaxLifetime"
	defaultConnMaxIdleKey  = "etc.cache.redis.connMaxIdleTime"
	// 超时配置键
	defaultDialTimeoutKey  = "etc.cache.redis.dialTimeout"
	defaultReadTimeoutKey  = "etc.cache.redis.readTimeout"
	defaultWriteTimeoutKey = "etc.cache.redis.writeTimeout"
)

type Option func(o *options)

type options struct {
	// 部署模式：single（单机）、sentinel（哨兵）、cluster（集群）
	// 默认为 single
	mode Mode

	// 客户端连接地址
	// 单机模式：["127.0.0.1:6379"]
	// 哨兵模式：哨兵节点地址列表
	// 集群模式：集群节点地址列表
	addrs []string

	// 数据库号（集群模式不支持）
	// 内建客户端配置，默认为0
	db int

	// 用户名
	// 内建客户端配置，默认为空
	username string

	// 密码
	// 内建客户端配置，默认为空
	password string

	// 客户端证书
	certFile string

	// 客户端密钥
	keyFile string

	// CA证书
	caFile string

	// 最大重试次数
	// 内建客户端配置，默认为3次
	maxRetries int

	// 客户端
	// 外部客户端配置，存在外部客户端时，优先使用外部客户端，默认为nil
	client redis.UniversalClient

	// 前缀
	// key前缀，默认为cache
	prefix string

	// 空值，默认为cache@nil
	nilValue string

	// 空值过期时间，默认为10s
	nilExpiration time.Duration

	// 最小过期时间，默认为1h
	minExpiration time.Duration

	// 最大过期时间，默认为24h
	maxExpiration time.Duration

	// ============================================================================
	// 哨兵模式配置
	// ============================================================================

	// 哨兵模式主节点名称
	masterName string

	// ============================================================================
	// 连接池配置
	// ============================================================================

	// 连接池大小，默认为 10 * runtime.GOMAXPROCS
	poolSize int

	// 最小空闲连接数
	minIdleConns int

	// 连接最大生命周期
	connMaxLifetime time.Duration

	// 连接最大空闲时间
	connMaxIdleTime time.Duration

	// ============================================================================
	// 超时配置
	// ============================================================================

	// 建立连接超时时间
	dialTimeout time.Duration

	// 读取超时时间
	readTimeout time.Duration

	// 写入超时时间
	writeTimeout time.Duration
}

func defaultOptions() *options {
	// 获取部署模式
	modeStr := etc.Get(defaultModeKey, string(ModeSingle)).String()
	mode := Mode(modeStr)

	// 根据模式获取地址
	var addrs []string
	switch mode {
	case ModeSentinel:
		addrs = etc.Get(defaultSentinelsKey, []string{defaultAddr}).Strings()
	case ModeCluster:
		addrs = etc.Get(defaultClusterAddrsKey, []string{defaultAddr}).Strings()
	default:
		addrs = etc.Get(defaultAddrsKey, []string{defaultAddr}).Strings()
	}

	return &options{
		mode:            mode,
		addrs:           addrs,
		db:              etc.Get(defaultDBKey, defaultDB).Int(),
		username:        etc.Get(defaultUsernameKey).String(),
		password:        etc.Get(defaultPasswordKey).String(),
		certFile:        etc.Get(defaultCertFileKey).String(),
		keyFile:         etc.Get(defaultKeyFileKey).String(),
		caFile:          etc.Get(defaultCAFileKey).String(),
		maxRetries:      etc.Get(defaultMaxRetriesKey, defaultMaxRetries).Int(),
		prefix:          etc.Get(defaultPrefixKey, defaultPrefix).String(),
		nilValue:        etc.Get(defaultNilValueKey, defaultNilValue).String(),
		nilExpiration:   etc.Get(defaultNilExpirationKey, defaultNilExpiration).Duration(),
		minExpiration:   etc.Get(defaultMinExpirationKey, defaultMinExpiration).Duration(),
		maxExpiration:   etc.Get(defaultMaxExpirationKey, defaultMaxExpiration).Duration(),
		masterName:      etc.Get(defaultMasterNameKey).String(),
		poolSize:        etc.Get(defaultPoolSizeKey).Int(),
		minIdleConns:    etc.Get(defaultMinIdleConnsKey).Int(),
		connMaxLifetime: etc.Get(defaultConnMaxLifeKey).Duration(),
		connMaxIdleTime: etc.Get(defaultConnMaxIdleKey).Duration(),
		dialTimeout:     etc.Get(defaultDialTimeoutKey).Duration(),
		readTimeout:     etc.Get(defaultReadTimeoutKey).Duration(),
		writeTimeout:    etc.Get(defaultWriteTimeoutKey).Duration(),
	}
}

// WithAddrs 设置连接地址
func WithAddrs(addrs ...string) Option {
	return func(o *options) { o.addrs = addrs }
}

// WithDB 设置数据库号
func WithDB(db int) Option {
	return func(o *options) { o.db = db }
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *options) { o.username = username }
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(o *options) { o.password = password }
}

// WithCredentials 设置证书、密钥、CA证书
func WithCredentials(certFile, keyFile, caFile string) Option {
	return func(o *options) { o.certFile, o.keyFile, o.caFile = certFile, keyFile, caFile }
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(o *options) { o.maxRetries = maxRetries }
}

// WithClient 设置外部客户端
func WithClient(client redis.UniversalClient) Option {
	return func(o *options) { o.client = client }
}

// WithPrefix 设置前缀
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// WithNilValue 设置空值
func WithNilValue(nilValue string) Option {
	return func(o *options) { o.nilValue = nilValue }
}

// WithNilExpiration 设置空值过期时间
func WithNilExpiration(nilExpiration time.Duration) Option {
	return func(o *options) { o.nilExpiration = nilExpiration }
}

// WithMinExpiration 设置最小过期时间
func WithMinExpiration(minExpiration time.Duration) Option {
	return func(o *options) { o.minExpiration = minExpiration }
}

// WithMaxExpiration 设置最大过期时间
func WithMaxExpiration(maxExpiration time.Duration) Option {
	return func(o *options) { o.maxExpiration = maxExpiration }
}

// ============================================================================
// 高可用配置选项
// ============================================================================

// WithMode 设置部署模式
func WithMode(mode Mode) Option {
	return func(o *options) { o.mode = mode }
}

// WithMasterName 设置哨兵模式主节点名称
func WithMasterName(masterName string) Option {
	return func(o *options) { o.masterName = masterName }
}

// ============================================================================
// 连接池配置选项
// ============================================================================

// WithPoolSize 设置连接池大小
func WithPoolSize(poolSize int) Option {
	return func(o *options) { o.poolSize = poolSize }
}

// WithMinIdleConns 设置最小空闲连接数
func WithMinIdleConns(minIdleConns int) Option {
	return func(o *options) { o.minIdleConns = minIdleConns }
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(o *options) { o.connMaxLifetime = connMaxLifetime }
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return func(o *options) { o.connMaxIdleTime = connMaxIdleTime }
}

// ============================================================================
// 超时配置选项
// ============================================================================

// WithDialTimeout 设置建立连接超时时间
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(o *options) { o.dialTimeout = dialTimeout }
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(o *options) { o.readTimeout = readTimeout }
}

// WithWriteTimeout 设置写入超时时间
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(o *options) { o.writeTimeout = writeTimeout }
}

// WithTimeouts 设置所有超时时间
func WithTimeouts(dialTimeout, readTimeout, writeTimeout time.Duration) Option {
	return func(o *options) {
		o.dialTimeout = dialTimeout
		o.readTimeout = readTimeout
		o.writeTimeout = writeTimeout
	}
}
