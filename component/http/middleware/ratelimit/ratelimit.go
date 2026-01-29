// Package ratelimit 提供 HTTP 请求限流中间件
// 支持多种限流策略：固定窗口、滑动窗口、令牌桶
// 支持多种存储后端：内存、Redis
package ratelimit

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/dawnsgo/dawn/component/http"
	"github.com/dawnsgo/dawn/etc"
	"github.com/dawnsgo/dawn/log"
	"github.com/gofiber/fiber/v3"
)

// ============================================================================
// 配置
// ============================================================================

// Config 限流配置
type Config struct {
	// Limit 限制次数
	Limit int `json:"limit"`
	// Window 时间窗口
	Window time.Duration `json:"window"`
	// KeyFunc 键生成函数（用于确定限流维度）
	KeyFunc KeyFunc
	// Storage 存储后端
	Storage Storage
	// SkipFunc 跳过检查函数
	SkipFunc SkipFunc
	// ErrorHandler 超限错误处理函数
	ErrorHandler ErrorHandler
	// Headers 是否添加限流响应头
	Headers bool
}

// KeyFunc 限流键生成函数
type KeyFunc func(ctx http.Context) string

// SkipFunc 跳过限流检查函数
type SkipFunc func(ctx http.Context) bool

// ErrorHandler 限流错误处理函数
type ErrorHandler func(ctx http.Context, remaining int, resetTime time.Time) error

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Limit:   100,
		Window:  time.Minute,
		Headers: true,
		KeyFunc: func(ctx http.Context) string {
			return ctx.IP()
		},
		ErrorHandler: func(ctx http.Context, remaining int, resetTime time.Time) error {
			ctx.Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
			return ctx.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"code":    429,
				"message": "请求过于频繁，请稍后重试",
			})
		},
	}
}

// ============================================================================
// 中间件
// ============================================================================

// New 创建限流中间件
func New(config ...Config) http.Handler {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = mergeConfig(cfg, config[0])
	}

	// 如果没有提供存储，使用内存存储
	if cfg.Storage == nil {
		cfg.Storage = NewMemoryStorage()
	}

	return func(ctx http.Context) error {
		// 检查是否跳过
		if cfg.SkipFunc != nil && cfg.SkipFunc(ctx) {
			return ctx.Next()
		}

		// 获取限流键
		key := "ratelimit:" + cfg.KeyFunc(ctx)

		// 执行限流检查
		result, err := cfg.Storage.Allow(key, cfg.Limit, cfg.Window)
		if err != nil {
			log.Warnf("ratelimit storage error: %v", err)
			// 存储错误时放行请求
			return ctx.Next()
		}

		// 设置响应头
		if cfg.Headers {
			ctx.Set("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
			ctx.Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			ctx.Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
		}

		// 检查是否超限
		if !result.Allowed {
			return cfg.ErrorHandler(ctx, result.Remaining, result.ResetAt)
		}

		return ctx.Next()
	}
}

// NewFromConfig 从配置文件创建限流中间件
func NewFromConfig(configKey string) http.Handler {
	cfg := DefaultConfig()

	var fileCfg struct {
		Limit  int    `json:"limit"`
		Window string `json:"window"`
	}

	if err := etc.Get(configKey).Scan(&fileCfg); err == nil {
		cfg.Limit = fileCfg.Limit
		if d, err := time.ParseDuration(fileCfg.Window); err == nil {
			cfg.Window = d
		}
	}

	return New(cfg)
}

// mergeConfig 合并配置
func mergeConfig(base, override Config) Config {
	if override.Limit > 0 {
		base.Limit = override.Limit
	}
	if override.Window > 0 {
		base.Window = override.Window
	}
	if override.KeyFunc != nil {
		base.KeyFunc = override.KeyFunc
	}
	if override.Storage != nil {
		base.Storage = override.Storage
	}
	if override.SkipFunc != nil {
		base.SkipFunc = override.SkipFunc
	}
	if override.ErrorHandler != nil {
		base.ErrorHandler = override.ErrorHandler
	}
	base.Headers = override.Headers
	return base
}

// ============================================================================
// 存储接口
// ============================================================================

// Storage 限流存储接口
type Storage interface {
	// Allow 检查是否允许请求
	Allow(key string, limit int, window time.Duration) (*Result, error)
	// Close 关闭存储
	Close() error
}

// Result 限流检查结果
type Result struct {
	// Allowed 是否允许
	Allowed bool
	// Remaining 剩余次数
	Remaining int
	// ResetAt 重置时间
	ResetAt time.Time
	// Current 当前计数
	Current int
}

// ============================================================================
// 内存存储实现
// ============================================================================

type memoryEntry struct {
	count    int
	expireAt time.Time
}

type memoryStorage struct {
	entries map[string]*memoryEntry
	mu      sync.RWMutex
	done    chan struct{}
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() Storage {
	s := &memoryStorage{
		entries: make(map[string]*memoryEntry),
		done:    make(chan struct{}),
	}

	// 启动清理协程
	go s.cleanup()

	return s
}

func (s *memoryStorage) Allow(key string, limit int, window time.Duration) (*Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	resetAt := now.Add(window)

	entry, exists := s.entries[key]
	if !exists || now.After(entry.expireAt) {
		// 新条目或已过期
		s.entries[key] = &memoryEntry{
			count:    1,
			expireAt: resetAt,
		}
		return &Result{
			Allowed:   true,
			Remaining: limit - 1,
			ResetAt:   resetAt,
			Current:   1,
		}, nil
	}

	// 增加计数
	entry.count++
	remaining := limit - entry.count
	if remaining < 0 {
		remaining = 0
	}

	return &Result{
		Allowed:   entry.count <= limit,
		Remaining: remaining,
		ResetAt:   entry.expireAt,
		Current:   entry.count,
	}, nil
}

func (s *memoryStorage) Close() error {
	close(s.done)
	return nil
}

func (s *memoryStorage) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, entry := range s.entries {
				if now.After(entry.expireAt) {
					delete(s.entries, key)
				}
			}
			s.mu.Unlock()
		}
	}
}

// ============================================================================
// Redis 存储实现
// ============================================================================

// RedisClient Redis 客户端接口
type RedisClient interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Get(ctx context.Context, key string) (string, error)
}

type redisStorage struct {
	client RedisClient
}

// NewRedisStorage 创建 Redis 存储
func NewRedisStorage(client RedisClient) Storage {
	return &redisStorage{client: client}
}

func (s *redisStorage) Allow(key string, limit int, window time.Duration) (*Result, error) {
	ctx := context.Background()

	// 增加计数
	count, err := s.client.Incr(ctx, key)
	if err != nil {
		return nil, err
	}

	// 首次设置过期时间
	if count == 1 {
		s.client.Expire(ctx, key, window)
	}

	// 获取剩余 TTL
	ttl, err := s.client.TTL(ctx, key)
	if err != nil {
		ttl = window
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return &Result{
		Allowed:   int(count) <= limit,
		Remaining: remaining,
		ResetAt:   time.Now().Add(ttl),
		Current:   int(count),
	}, nil
}

func (s *redisStorage) Close() error {
	return nil
}

// ============================================================================
// 便捷函数
// ============================================================================

// IPRateLimit 基于 IP 的限流
func IPRateLimit(limit int, window time.Duration) http.Handler {
	return New(Config{
		Limit:  limit,
		Window: window,
		KeyFunc: func(ctx http.Context) string {
			return "ip:" + ctx.IP()
		},
	})
}

// PathRateLimit 基于路径的限流
func PathRateLimit(limit int, window time.Duration) http.Handler {
	return New(Config{
		Limit:  limit,
		Window: window,
		KeyFunc: func(ctx http.Context) string {
			return "path:" + ctx.Path() + ":" + ctx.IP()
		},
	})
}

// UserRateLimit 基于用户 ID 的限流（需要提供获取用户 ID 的函数）
func UserRateLimit(limit int, window time.Duration, getUserID func(ctx http.Context) string) http.Handler {
	return New(Config{
		Limit:  limit,
		Window: window,
		KeyFunc: func(ctx http.Context) string {
			uid := getUserID(ctx)
			if uid == "" {
				return "user:anonymous:" + ctx.IP()
			}
			return "user:" + uid
		},
	})
}
