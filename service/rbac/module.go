package rbac

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"sitehub/common/config"
	"sitehub/common/infrastructure/smtp"
	"sitehub/common/pkg/session"
	"sitehub/common/pkg/token"
	settingsrepo "sitehub/common/settings/contract"
	rbacservice "sitehub/service/rbac/business"
	"sitehub/service/rbac/contract"
	"sitehub/service/rbac/model/impl"
	mysql "sitehub/service/rbac/model/mysql"
)

type Module struct {
	Members     *rbacservice.MemberService
	Users       rbacservice.SysUserService
	Roles       rbacservice.SysRoleService
	Menus       rbacservice.SysMenuService
	Systems     *rbacservice.SysSystemService
	Permissions rbacservice.SysPermissionService
	Captcha     rbacservice.SliderCaptchaService
	Email       *rbacservice.EmailSettingsService
	Sessions    session.Manager
	Tokens      token.Manager
	Blacklist   contract.Blacklist
}

func New(db *gorm.DB, cache *redis.Client, cfg *config.Config, configs settingsrepo.AtomicStore) *Module {
	users, roles, menus := mysql.NewSysUserRepositoryImpl(db), mysql.NewSysRoleRepositoryImpl(db), mysql.NewSysMenuRepositoryImpl(db)
	blacklist := mysql.NewIPBlacklistRepository(db)
	tokenConfig := token.DefaultConfig()
	if cfg.OnceToken.ExpireTime > 0 {
		tokenConfig.ExpireTime = cfg.OnceToken.ExpireTime
	}
	return &Module{
		Members: rbacservice.NewMemberService(mysql.NewMemberRepository(db), cache, cfg),
		Users:   rbacservice.NewSysUserService(users, mysql.NewSysLogRepositoryImpl(db), impl.NewRedisCacheRepo(cache)),
		Roles:   rbacservice.NewSysRoleService(roles), Menus: rbacservice.NewSysMenuService(menus),
		Systems:     rbacservice.NewSysSystemService(db),
		Permissions: rbacservice.NewSysPermissionService(users, roles, menus),
		Captcha:     rbacservice.NewSliderCaptchaService(cache), Email: rbacservice.NewEmailSettingsService(configs, cfg),
		Sessions: session.NewRedisSession(cache, &session.Config{CookieName: cfg.Session.CookieName, MaxAge: cfg.Session.MaxAge, Path: cfg.Session.Path, Domain: cfg.Session.Domain, Secure: cfg.Session.Secure, HttpOnly: cfg.Session.HttpOnly}),
		Tokens:   token.NewRedisTokenManager(cache, tokenConfig),
		// Resolve the current SMTP runtime when sending, after startup configuration
		// has loaded; capturing GetEmailServer here would retain an uninitialized sender.
		Blacklist: rbacservice.NewBlacklistService(blacklist, smtp.SendEmail, cfg.Notifications.BlacklistEmail),
	}
}
