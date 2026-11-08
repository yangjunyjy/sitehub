package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"strings"

	"sitehub/common/config"
	settingsrepository "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	rbacrequest "sitehub/service/rbac/dto/request"
)

const emailSettingsKey = "system.email"

var ErrInvalidEmailSettings = errors.New("邮件配置无效")

// EmailSettingsView deliberately excludes the SMTP password from every response.
type EmailSettingsView struct {
	SMTPHost    string `json:"smtpHost"`
	SMTPPort    int    `json:"smtpPort"`
	FromEmail   string `json:"fromEmail"`
	FromName    string `json:"fromName"`
	Subject     string `json:"subject"`
	HasPassword bool   `json:"hasPassword"`
	Source      string `json:"source"`
}

type EmailSettingsService struct {
	repo settingsrepository.AtomicStore
	cfg  *config.Config
}

func NewEmailSettingsService(repo settingsrepository.AtomicStore, cfg *config.Config) *EmailSettingsService {
	return &EmailSettingsService{repo: repo, cfg: cfg}
}

func (s *EmailSettingsService) read() (config.EmailConfig, string, error) {
	value := s.cfg.EmailServer
	row, err := s.repo.GetByKey(emailSettingsKey)
	if err != nil {
		return value, "", err
	}
	if row == nil {
		return value, "config", nil
	}
	if err := json.Unmarshal([]byte(row.Value), &value.SmtpConfig); err != nil {
		return value, "", errors.New("已保存的邮件配置无法读取")
	}
	return value, "database", nil
}

// RuntimeConfig is read before each send. No global sender is swapped while mail is in flight.
func (s *EmailSettingsService) RuntimeConfig() (config.EmailConfig, error) {
	value, _, err := s.read()
	return value, err
}

func emailSettingsView(value config.SmtpConfig, source string) EmailSettingsView {
	return EmailSettingsView{SMTPHost: value.SMTPHost, SMTPPort: value.SMTPPort, FromEmail: value.FromEmail,
		FromName: value.FromName, Subject: value.Subject, HasPassword: value.Password != "", Source: source}
}

func (s *EmailSettingsService) Get() (EmailSettingsView, error) {
	value, source, err := s.read()
	return emailSettingsView(value.SmtpConfig, source), err
}

func validateEmailSettings(value config.SmtpConfig) error {
	invalid := func(reason string) error { return fmt.Errorf("%w：%s", ErrInvalidEmailSettings, reason) }
	if value.SMTPHost == "" || len(value.SMTPHost) > 253 || strings.ContainsAny(value.SMTPHost, " /\\\t\r\n") || (strings.Contains(value.SMTPHost, ":") && net.ParseIP(value.SMTPHost) == nil) {
		return invalid("SMTP 主机应填写域名或 IP，不包含协议和端口")
	}
	if value.SMTPPort < 1 || value.SMTPPort > 65535 {
		return invalid("SMTP 端口应在 1–65535 之间")
	}
	address, err := mail.ParseAddress(value.FromEmail)
	if err != nil || address.Address != value.FromEmail || strings.ContainsAny(value.FromEmail, "\r\n") {
		return invalid("请填写有效的发件邮箱")
	}
	if value.FromName == "" || value.Subject == "" || len(value.FromName) > 200 || len(value.Subject) > 300 || strings.ContainsAny(value.FromName+value.Subject, "\r\n") {
		return invalid("发件名称和邮件主题不能为空，且不能包含换行")
	}
	if value.Password == "" || value.Password == "********" {
		return invalid("请填写 SMTP 密码或授权码")
	}
	return nil
}

func (s *EmailSettingsService) Save(req rbacrequest.UpdateEmailSettingsRequest) (EmailSettingsView, error) {
	initial, err := json.Marshal(s.cfg.EmailServer.SmtpConfig)
	if err != nil {
		return EmailSettingsView{}, err
	}
	var saved config.SmtpConfig
	err = s.repo.MutateValue(settingsmodel.SystemConfig{Key: emailSettingsKey, Value: string(initial), Category: settingsmodel.ConfigCategoryEmail,
		ConfigType: settingsmodel.ConfigTypeJSON, Label: "邮件服务", IsPublic: false, IsEditable: true}, func(current string) (string, error) {
		// The locked snapshot owns the retained password. SMTP I/O never runs in this transaction.
		if err := json.Unmarshal([]byte(current), &saved); err != nil {
			return "", errors.New("已保存的邮件配置无法读取")
		}
		saved.SMTPHost = strings.TrimSpace(req.SMTPHost)
		saved.SMTPPort = req.SMTPPort
		saved.FromEmail = strings.TrimSpace(req.FromEmail)
		saved.FromName = strings.TrimSpace(req.FromName)
		saved.Subject = strings.TrimSpace(req.Subject)
		if req.Password != nil {
			saved.Password = *req.Password
		}
		if err := validateEmailSettings(saved); err != nil {
			return "", err
		}
		data, err := json.Marshal(saved)
		return string(data), err
	})
	if err != nil {
		return EmailSettingsView{}, err
	}
	return emailSettingsView(saved, "database"), nil
}
