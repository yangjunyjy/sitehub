package services

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"sitehub/common/pkg/logger"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

type BlacklistService struct {
	repo        rbacrepository.IPBlacklistRepository
	emailSender func(to, subject, html string) error
	adminEmail  string
}

func NewBlacklistService(repo rbacrepository.IPBlacklistRepository, emailSender func(to, subject, html string) error, adminEmail string) *BlacklistService {
	return &BlacklistService{
		repo:        repo,
		emailSender: emailSender,
		adminEmail:  strings.TrimSpace(adminEmail),
	}
}

// AddToBlacklist 添加IP到黑名单并发送邮件通知
func (s *BlacklistService) AddToBlacklist(ip, reason string) error {
	// 检查IP是否已在黑名单中
	existing, err := s.repo.FindByIP(ip)
	if err == nil && existing != nil {
		logger.Warn("IP已在黑名单中", logger.String("ip", ip))
		return fmt.Errorf("IP %s 已在黑名单中", ip)
	}

	// 添加到黑名单
	if err := s.repo.Create(ip, reason); err != nil {
		logger.Error("添加IP到黑名单失败", logger.Err("error", err), logger.String("ip", ip))
		return fmt.Errorf("添加IP到黑名单失败: %v", err)
	}

	// 异步发送邮件通知
	go s.sendNotificationEmail(ip, reason)

	logger.Info("IP已添加到黑名单", logger.String("ip", ip), logger.String("reason", reason))
	return nil
}

// GetBlacklist 获取黑名单列表
func (s *BlacklistService) GetBlacklist(offset, limit int) ([]rbacentity.IPBlacklist, error) {
	return s.repo.List(offset, limit)
}

// GetByID 根据ID查询黑名单
func (s *BlacklistService) GetByID(id uint) (*rbacentity.IPBlacklist, error) {
	return s.repo.FindByID(id)
}

// GetByIP 根据IP查询黑名单
func (s *BlacklistService) GetByIP(ip string) (*rbacentity.IPBlacklist, error) {
	return s.repo.FindByIP(ip)
}

// GetBlockedBlacklist 获取已封禁的IP
func (s *BlacklistService) GetBlockedBlacklist() ([]rbacentity.IPBlacklist, error) {
	return s.repo.FindByIsBlocked(true)
}

// BlockedIPs is the cache-facing capability. Persistence entities stay in rbac.
func (s *BlacklistService) BlockedIPs() ([]string, error) {
	rows, err := s.repo.FindByIsBlocked(true)
	if err != nil {
		return nil, err
	}
	ips := make([]string, 0, len(rows))
	for _, row := range rows {
		ips = append(ips, row.IPAddress)
	}
	return ips, nil
}

// RemoveFromBlacklist 从黑名单中移除IP
func (s *BlacklistService) RemoveFromBlacklist(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		logger.Error("从黑名单移除IP失败", logger.Err("error", err), logger.Uint("id", id))
		return fmt.Errorf("从黑名单移除IP失败: %v", err)
	}

	logger.Info("IP已从黑名单移除", logger.Uint("id", id))
	return nil
}

// sendNotificationEmail 发送黑名单通知邮件
func (s *BlacklistService) sendNotificationEmail(ip, reason string) {
	if strings.TrimSpace(s.adminEmail) == "" {
		return
	}
	if s.emailSender == nil {
		logger.Warn("邮件发送器未初始化，跳过邮件通知")
		return
	}

	// 构建邮件内容
	subject := fmt.Sprintf("【安全警报】IP %s 因异常访问被加入黑名单", ip)
	htmlContent := s.buildHTMLNotificationEmail(ip, reason)

	// 发送邮件
	if err := s.emailSender(s.adminEmail, subject, htmlContent); err != nil {
		logger.Error("发送黑名单通知邮件失败", logger.Err("error", err), logger.String("to", s.adminEmail))
	} else {
		logger.Info("黑名单通知邮件发送成功", logger.String("to", s.adminEmail), logger.String("ip", ip))
	}
}

// buildHTMLNotificationEmail 构建HTML格式的通知邮件
func (s *BlacklistService) buildHTMLNotificationEmail(ip, reason string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>黑名单通知</title>
</head>
<body>
    <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
        <h2 style="color: #ff0000;">【安全警报】IP被加入黑名单</h2>
        <p>尊敬的管理员，您好！</p>
        <p>系统检测到以下IP地址存在异常访问行为，已自动将其加入黑名单：</p>
        <div style="background-color: #f5f5f5; padding: 15px; border-left: 4px solid #ff0000; margin: 20px 0;">
            <p><strong>IP地址：</strong>%s</p>
            <p><strong>封禁原因：</strong>%s</p>
        </div>
        <p>请及时关注服务器访问日志，排查是否存在安全风险。</p>
        <hr>
        <p style="color: #999; font-size: 12px;">此为系统自动发送邮件，请勿直接回复</p>
    </div>
</body>
</html>`, ip, reason)
}

// buildTextNotificationEmail 构建纯文本格式的通知邮件
func (s *BlacklistService) buildTextNotificationEmail(ip, reason string) string {
	return fmt.Sprintf(`尊敬的管理员，您好！

系统检测到以下IP地址存在异常访问行为，已自动将其加入黑名单：

IP地址：%s
封禁原因：%s

请及时关注服务器访问日志，排查是否存在安全风险。

此为系统自动发送邮件，请勿直接回复`, ip, reason)
}

// CheckIPBlocked 检查IP是否在黑名单中
func (s *BlacklistService) CheckIPBlocked(ip string) (bool, error) {
	blacklist, err := s.repo.FindByIP(ip)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return blacklist.IsBlocked, nil
}

// UnblockIP 解封IP
func (s *BlacklistService) UnblockIP(id uint) error {
	return s.RemoveFromBlacklist(id)
}

// GetBlacklistStats 获取黑名单统计信息
func (s *BlacklistService) GetBlacklistStats() (map[string]interface{}, error) {
	blockedIPs, err := s.repo.FindByIsBlocked(true)
	if err != nil {
		return nil, err
	}

	allIPs, err := s.repo.List(0, 10000) // 获取所有IP
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_count":     len(allIPs),
		"blocked_count":   len(blockedIPs),
		"unblocked_count": len(allIPs) - len(blockedIPs),
	}

	return stats, nil
}
