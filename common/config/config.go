package config

import (
	"time"
)

// Config holds shared business settings assembled by each service context.
// Network, database and discovery settings belong to the service configuration.
type Config struct {
	Auth          AuthConfig         `yaml:"auth" env:"AUTH"`
	Session       SessionConfig      `yaml:"session" env:"SESSION"`
	OnceToken     OnceTokenConfig    `yaml:"onceToken" env:"ONCE_TOKEN"`
	Log           LogConfig          `yaml:"log" env:"LOG"`
	EmailServer   EmailConfig        `yaml:"emailServer" json:"emailServer"`
	Upload        UploadConfig       `yaml:"upload" env:"UPLOAD"`
	Notifications NotificationConfig `yaml:"notifications"`
}

type NotificationConfig struct {
	BlacklistEmail string `yaml:"blacklistEmail"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string        `yaml:"jwtSecret" env:"JWT_SECRET"`
	JWTExpire        time.Duration `yaml:"jwtExpire" env:"JWT_EXPIRE" env-default:"24h"`
	JWTRefreshExpire time.Duration `yaml:"jwtRefreshExpire" env:"JWT_REFRESH_EXPIRE" env-default:"168h"`
	BcryptCost       int           `yaml:"bcryptCost" env:"BCRYPT_COST" env-default:"10"`
	Issuer           string        `yaml:"issuer" env:"ISSUER" env-default:"sitehub"`

	Password struct {
		MinLength    int  `yaml:"min_length" env:"MIN_LENGTH" env-default:"6"`
		RequireUpper bool `yaml:"require_upper" env:"REQUIRE_UPPER" env-default:"true"`
		RequireLower bool `yaml:"require_lower" env:"REQUIRE_LOWER" env-default:"true"`
		RequireDigit bool `yaml:"require_digit" env:"REQUIRE_DIGIT" env-default:"true"`
	} `yaml:"password"`

	// TOTP配置（基于时间的动态Token）
	TOTP struct {
		Enabled    bool   `yaml:"enabled" env:"TOTP_ENABLED" env-default:"false"`
		Secret     string `yaml:"secret" env:"TOTP_SECRET"`                          // 全局共享密钥
		TimeStep   int    `yaml:"timeStep" env:"TOTP_TIME_STEP" env-default:"30"`    // 时间步长（秒）
		WindowSize int    `yaml:"windowSize" env:"TOTP_WINDOW_SIZE" env-default:"1"` // 验证窗口大小
	} `yaml:"totp"`
}

// SessionConfig Session配置
type SessionConfig struct {
	CookieName string `yaml:"cookieName" env:"COOKIE_NAME" env-default:"session_id"`
	MaxAge     int    `yaml:"maxAge" env:"MAX_AGE" env-default:"86400"`
	Path       string `yaml:"path" env:"PATH" env-default:"/"`
	Domain     string `yaml:"domain" env:"DOMAIN" env-default:""`
	Secure     bool   `yaml:"secure" env:"SECURE" env-default:"false"`
	HttpOnly   bool   `yaml:"httpOnly" env:"HTTP_ONLY" env-default:"true"`
}

// OnceTokenConfig 一次性Token配置
type OnceTokenConfig struct {
	ExpireTime int `yaml:"expireTime" env:"EXPIRE_TIME" env-default:"300"`
}

// LogConfig 日志配置
type LogConfig struct {
	Devolpment    bool   `yaml:"devolpment" json:"devolpment"`       //运行模式
	Level         string `yaml:"level" json:"level"`                 //日志级别
	Encoding      string `yaml:"encoding" json:"encoding"`           //输出格式，console,json
	Directory     string `yaml:"directory" json:"directory"`         //日志目录
	FileName      string `yaml:"fileName" json:"fileName"`           //日志名
	MaxSize       uint   `yaml:"maxSize" json:"maxSize"`             //日志文件最大大小
	MaxBackups    uint   `yaml:"maxBackups" json:"maxBackups"`       //保留旧文件最大个数
	MaxAge        uint   `yaml:"maxAge" json:"maxAge"`               //保留旧文件最大天数
	Compress      bool   `yaml:"compress" json:"compress"`           //是否压缩旧文件
	EnableConsole bool   `yaml:"enableConsole" json:"enableConsole"` //是否启用控制台输出
	FileConsole   bool   `yaml:"fileConsole" json:"fileConsole"`     //是否启用文件输出
	EnableCaller  bool   `yaml:"enableCaller" json:"enableCaller"`   //是否添加调用者信息
}

func (config *LogConfig) SetDefault() {
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Encoding == "" {
		config.Encoding = "json"
	}
	if config.FileName == "" {
		config.FileName = "app.log"
	}
	if config.MaxSize == 0 {
		config.MaxSize = 100 // 100MB
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 30 // 保留30个备份
	}
	if config.MaxAge == 0 {
		config.MaxAge = 90 // 保留90天
	}
}

// SmtpConfig defines the SMTP server and sender settings.
type SmtpConfig struct {
	SMTPHost  string `yaml:"smtpHost" json:"smtpHost"`   // SMTP服务器地址
	SMTPPort  int    `yaml:"smtpPort" json:"smtpPort"`   // SMTP端口
	FromEmail string `yaml:"fromEmail" json:"fromEmail"` // 发件人邮箱
	FromName  string `yaml:"fromName" json:"fromName"`   // 发件人名称
	Password  string `yaml:"password" json:"password"`   // 授权码/密码
	Subject   string `yaml:"subject" json:"subject"`     // 邮件主题
	PoolSize  int    `yaml:"poolSize" json:"poolSize"`   // 连接池大小
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SmtpConfig             `yaml:"smtpConfig" json:"smtpServerConfig"`
	VerificationCodeConfig `yaml:"verificationCodeConfig" json:"codeConfig"`
}

type VerificationCodeConfig struct {
	Length       int           `yaml:"length" env:"LENGTH" env-default:"6"`            // 验证码长度
	Charset      string        `yaml:"charset" env:"CHARSET" env-default:"0123456789"` // 字符集
	ExpiresIn    time.Duration `yaml:"expiresIn" env:"EXPIRES_IN" env-default:"5m"`    // 过期时间
	HTMLTemplate string        `yaml:"htmlTemplate" env:"HTML_TEMPLATE"`               // HTML模板
	TextTemplate string        `yaml:"textTemplate" env:"TEXT_TEMPLATE"`               // 纯文本模板
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	// 上传方式: local 或 remote
	Type string `yaml:"type" env:"TYPE" env-default:"local"`

	// 本地上传配置
	Local struct {
		UploadPath  string   `yaml:"uploadPath" env:"UPLOAD_PATH" env-default:"./uploads"` // 上传目录
		AllowedExts []string `yaml:"allowedExts" env:"ALLOWED_EXTS" env-separator:","`     // 允许的文件扩展名
		MaxSize     int64    `yaml:"maxSize" env:"MAX_SIZE" env-default:"10485760"`        // 最大文件大小(10MB)
	} `yaml:"local"`

	// 远程上传配置 (如 OSS, COS, 七牛云等)
	Remote struct {
		Provider  string `yaml:"provider" env:"PROVIDER" env-default:"oss"` // 服务提供商: oss, minio
		Endpoint  string `yaml:"endpoint" env:"ENDPOINT"`                   // 服务端点
		AccessKey string `yaml:"accessKey" env:"ACCESS_KEY"`                // 访问密钥
		SecretKey string `yaml:"secretKey" env:"SECRET_KEY"`                // 密钥
		Bucket    string `yaml:"bucket" env:"BUCKET"`                       // 存储桶名称
		Region    string `yaml:"region" env:"REGION"`                       // 区域
		CDN       string `yaml:"cdn" env:"CDN"`                             // CDN域名
		PathStyle bool   `yaml:"pathStyle" env:"PATH_STYLE"`                // MinIO 通常为 true，OSS 通常为 false
	} `yaml:"remote"`

	// URL 配置
	URLPrefix string `yaml:"urlPrefix" env:"URL_PREFIX"` // URL前缀，用于拼接完整URL
}
