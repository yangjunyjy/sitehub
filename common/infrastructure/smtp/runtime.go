package smtp

import (
	"fmt"
	"sync"

	"sitehub/common/config"
)

var runtimeConfig struct {
	sync.RWMutex
	provider func() (config.EmailConfig, error)
}

// SetConfigProvider installs the read-only runtime configuration resolver at startup.
func SetConfigProvider(provider func() (config.EmailConfig, error)) {
	runtimeConfig.Lock()
	defer runtimeConfig.Unlock()
	runtimeConfig.provider = provider
}

func acquireSender() (*Sender, func(), error) {
	runtimeConfig.RLock()
	provider := runtimeConfig.provider
	runtimeConfig.RUnlock()
	if provider == nil {
		if globalEmailServer == nil {
			return nil, nil, fmt.Errorf("邮件服务未初始化")
		}
		return globalEmailServer, func() {}, nil
	}
	value, err := provider()
	if err != nil {
		return nil, nil, fmt.Errorf("读取邮件配置失败: %w", err)
	}
	sender, err := NewEmailSender(value.SmtpConfig, value.VerificationCodeConfig)
	if err != nil {
		return nil, nil, err
	}
	return sender, sender.Close, nil
}

// SendVerificationCode resolves one committed configuration snapshot for the entire send.
func SendVerificationCode(email, name string) (string, error) {
	sender, release, err := acquireSender()
	if err != nil {
		return "", err
	}
	defer release()
	return sender.SendVerificationCode(email, name)
}
