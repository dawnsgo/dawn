package log

import (
	"regexp"
	"strings"
	"sync"

	"github.com/dawnsgo/dawn/etc"
)

// ============================================================================
// 日志脱敏配置
// ============================================================================

// SanitizerConfig 日志脱敏配置
type SanitizerConfig struct {
	// Enabled 是否启用脱敏
	Enabled bool `json:"enabled"`
	// Fields 需要脱敏的字段名（如 password, token, secret）
	Fields []string `json:"fields"`
	// Patterns 需要脱敏的数据模式（如 phone, idcard, email）
	Patterns []string `json:"patterns"`
	// MaskChar 脱敏字符
	MaskChar string `json:"maskChar"`
	// PreserveLength 保留原始长度
	PreserveLength bool `json:"preserveLength"`
}

// DefaultSanitizerConfig 默认脱敏配置
func DefaultSanitizerConfig() *SanitizerConfig {
	return &SanitizerConfig{
		Enabled: false,
		Fields: []string{
			"password", "pwd", "passwd",
			"token", "accessToken", "refreshToken",
			"secret", "secretKey", "apiKey", "api_key",
			"credential", "credentials",
			"authorization",
		},
		Patterns: []string{"phone", "idcard", "email", "bankcard"},
		MaskChar: "*",
	}
}

// ============================================================================
// 内置脱敏模式
// ============================================================================

// PatternRule 脱敏模式规则
type PatternRule struct {
	Name    string
	Regex   *regexp.Regexp
	Replace func(match string) string
}

var builtinPatterns = map[string]*PatternRule{
	// 手机号：保留前3后4
	"phone": {
		Name:  "phone",
		Regex: regexp.MustCompile(`1[3-9]\d{9}`),
		Replace: func(match string) string {
			if len(match) >= 11 {
				return match[:3] + "****" + match[7:]
			}
			return "***"
		},
	},
	// 身份证号：保留前6后4
	"idcard": {
		Name:  "idcard",
		Regex: regexp.MustCompile(`[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`),
		Replace: func(match string) string {
			if len(match) >= 18 {
				return match[:6] + "********" + match[14:]
			}
			return "***"
		},
	},
	// 邮箱：保留首字母和域名
	"email": {
		Name:  "email",
		Regex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		Replace: func(match string) string {
			parts := strings.Split(match, "@")
			if len(parts) == 2 {
				local := parts[0]
				if len(local) > 1 {
					return local[:1] + "***@" + parts[1]
				}
				return "***@" + parts[1]
			}
			return "***"
		},
	},
	// 银行卡号：保留前6后4
	"bankcard": {
		Name:  "bankcard",
		Regex: regexp.MustCompile(`\d{13,19}`),
		Replace: func(match string) string {
			if len(match) >= 10 {
				return match[:6] + "****" + match[len(match)-4:]
			}
			return "***"
		},
	},
	// IP地址：保留前两段
	"ip": {
		Name:  "ip",
		Regex: regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
		Replace: func(match string) string {
			parts := strings.Split(match, ".")
			if len(parts) == 4 {
				return parts[0] + "." + parts[1] + ".*.*"
			}
			return "***"
		},
	},
}

// ============================================================================
// 日志脱敏器
// ============================================================================

// Sanitizer 日志脱敏器
type Sanitizer struct {
	config   *SanitizerConfig
	patterns []*PatternRule
	fieldRe  *regexp.Regexp
	mu       sync.RWMutex
}

var (
	globalSanitizer *Sanitizer
	sanitizerOnce   sync.Once
)

// GetSanitizer 获取全局脱敏器
func GetSanitizer() *Sanitizer {
	sanitizerOnce.Do(func() {
		globalSanitizer = NewSanitizer()
	})
	return globalSanitizer
}

// SetSanitizer 设置全局脱敏器
func SetSanitizer(s *Sanitizer) {
	globalSanitizer = s
}

// NewSanitizer 创建新的脱敏器
func NewSanitizer(configKey ...string) *Sanitizer {
	cfg := DefaultSanitizerConfig()

	key := "etc.log.sanitizer"
	if len(configKey) > 0 && configKey[0] != "" {
		key = configKey[0]
	}

	if err := etc.Get(key).Scan(cfg); err != nil {
		// 配置加载失败，使用默认配置
	}

	return NewSanitizerWithConfig(cfg)
}

// NewSanitizerWithConfig 使用配置创建脱敏器
func NewSanitizerWithConfig(cfg *SanitizerConfig) *Sanitizer {
	if cfg == nil {
		cfg = DefaultSanitizerConfig()
	}

	s := &Sanitizer{
		config:   cfg,
		patterns: make([]*PatternRule, 0),
	}

	// 加载内置模式
	for _, name := range cfg.Patterns {
		if rule, ok := builtinPatterns[name]; ok {
			s.patterns = append(s.patterns, rule)
		}
	}

	// 构建字段名匹配正则
	if len(cfg.Fields) > 0 {
		// 匹配格式：field=value, field:value, "field":"value", field="value"
		fieldPattern := `(?i)(` + strings.Join(cfg.Fields, "|") + `)\s*[=:]\s*["']?([^"',\s\}]+)["']?`
		s.fieldRe = regexp.MustCompile(fieldPattern)
	}

	return s
}

// Sanitize 对字符串进行脱敏处理
func (s *Sanitizer) Sanitize(msg string) string {
	if !s.config.Enabled {
		return msg
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := msg

	// 1. 字段名脱敏
	if s.fieldRe != nil {
		result = s.fieldRe.ReplaceAllStringFunc(result, func(match string) string {
			// 找到等号或冒号的位置
			idx := strings.IndexAny(match, "=:")
			if idx == -1 {
				return match
			}
			key := match[:idx]
			sep := string(match[idx])
			// 保留键名，替换值
			return key + sep + s.maskValue(match[idx+1:])
		})
	}

	// 2. 模式脱敏
	for _, rule := range s.patterns {
		result = rule.Regex.ReplaceAllStringFunc(result, rule.Replace)
	}

	return result
}

// maskValue 掩盖值
func (s *Sanitizer) maskValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)

	if len(value) == 0 {
		return value
	}

	maskChar := s.config.MaskChar
	if maskChar == "" {
		maskChar = "*"
	}

	if s.config.PreserveLength {
		return strings.Repeat(maskChar, len(value))
	}

	// 默认返回固定长度的掩码
	return strings.Repeat(maskChar, 6)
}

// AddPattern 添加自定义脱敏模式
func (s *Sanitizer) AddPattern(name string, regex *regexp.Regexp, replace func(string) string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.patterns = append(s.patterns, &PatternRule{
		Name:    name,
		Regex:   regex,
		Replace: replace,
	})
}

// AddField 添加需要脱敏的字段
func (s *Sanitizer) AddField(fields ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Fields = append(s.config.Fields, fields...)

	// 重新构建正则
	if len(s.config.Fields) > 0 {
		fieldPattern := `(?i)(` + strings.Join(s.config.Fields, "|") + `)\s*[=:]\s*["']?([^"',\s\}]+)["']?`
		s.fieldRe = regexp.MustCompile(fieldPattern)
	}
}

// Enable 启用脱敏
func (s *Sanitizer) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Enabled = true
}

// Disable 禁用脱敏
func (s *Sanitizer) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Enabled = false
}

// IsEnabled 是否启用
func (s *Sanitizer) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Enabled
}

// ============================================================================
// 便捷函数
// ============================================================================

// SanitizeMessage 对消息进行脱敏（使用全局脱敏器）
func SanitizeMessage(msg string) string {
	return GetSanitizer().Sanitize(msg)
}

// EnableSanitizer 启用全局脱敏器
func EnableSanitizer() {
	GetSanitizer().Enable()
}

// DisableSanitizer 禁用全局脱敏器
func DisableSanitizer() {
	GetSanitizer().Disable()
}
