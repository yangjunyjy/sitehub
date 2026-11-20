package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	errorx "sitehub/common/errorx"
	"sitehub/common/pkg/dberror"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

// SysMenuService 菜单服务
type SysMenuService interface {
	// List 列表
	List(page, pageSize int) ([]*rbacentity.SysMenu, int64, error)
	// GetTree 获取菜单树
	GetTree() ([]*rbacentity.SysMenu, error)
	// GetByID 获取详情
	GetByID(id uint64) (*rbacentity.SysMenu, error)
	// Create 创建
	Create(menu *rbacentity.SysMenu) error
	// Update 更新
	Update(menu *rbacentity.SysMenu) error
	// Delete 删除
	Delete(id uint64) error
	// GetUserMenus 获取用户菜单
	GetUserMenus(userID uint64) ([]*rbacentity.SysMenu, error)
	// 获取所有菜单
	GetAllMenus() ([]*rbacentity.SysMenu, error)
}

type sysMenuService struct {
	menuRepo rbacrepository.SysMenuRepository
}

func NewSysMenuService(menuRepo rbacrepository.SysMenuRepository) SysMenuService {
	return &sysMenuService{menuRepo: menuRepo}
}

func (s *sysMenuService) List(page, pageSize int) ([]*rbacentity.SysMenu, int64, error) {
	return s.menuRepo.List(page, pageSize)
}

func (s *sysMenuService) GetTree() ([]*rbacentity.SysMenu, error) {
	return s.menuRepo.GetMenuTree()
}

func (s *sysMenuService) GetByID(id uint64) (*rbacentity.SysMenu, error) {
	return s.menuRepo.FindByID(id)
}

func (s *sysMenuService) Create(menu *rbacentity.SysMenu) error {
	if err := prepareMenu(menu); err != nil {
		return err
	}
	if menu.MenuCode == "" {
		menu.MenuCode = "menu:" + uuid.NewString()
	}
	if err := checkMenuUniqueness(s.menuRepo, menu); err != nil {
		return err
	}
	return menuWriteError(s.menuRepo.Create(menu))
}

func (s *sysMenuService) Update(menu *rbacentity.SysMenu) error {
	if err := prepareMenu(menu); err != nil {
		return err
	}
	if menu.ID == 0 {
		return errorx.Invalid("无效的菜单ID")
	}
	// Read and preserve the current code under the same transaction/row lock
	// used by the write; an omitted code cannot erase a concurrent edit.
	return menuWriteError(s.menuRepo.InTransaction(func(repo rbacrepository.SysMenuRepository) error {
		current, err := repo.FindByIDForUpdate(menu.ID)
		if err != nil {
			return err
		}
		if menu.MenuCode == "" {
			menu.MenuCode = current.MenuCode
			if strings.TrimSpace(menu.MenuCode) == "" {
				menu.MenuCode = "menu:" + uuid.NewString()
			}
		}
		if err := checkMenuUniqueness(repo, menu); err != nil {
			return err
		}
		return repo.Update(menu)
	}))
}

func prepareMenu(menu *rbacentity.SysMenu) error {
	if menu == nil {
		return errorx.Invalid("菜单信息不能为空")
	}
	menu.MenuName = strings.TrimSpace(menu.MenuName)
	menu.SystemCode = strings.TrimSpace(menu.SystemCode)
	if menu.SystemCode == "" {
		menu.SystemCode = "admin"
	}
	if !systemCodePattern.MatchString(menu.SystemCode) {
		return errorx.Invalid("系统代码无效")
	}
	menu.Permission = strings.TrimSpace(menu.Permission)
	menu.MenuCode = strings.TrimSpace(menu.MenuCode)
	menu.Path = strings.TrimSpace(menu.Path)
	switch menu.PageType {
	case "", "internal":
		menu.PageType = "internal"
		menu.ExternalURL = ""
	case "external":
		menu.ExternalURL = strings.TrimSpace(menu.ExternalURL)
		address, err := url.Parse(menu.ExternalURL)
		if menu.MenuType != 2 || len(menu.ExternalURL) > 2048 || err != nil || address.Hostname() == "" || address.User != nil || (address.Scheme != "http" && address.Scheme != "https") || strings.ContainsAny(menu.ExternalURL, "\\ \t\r\n") {
			return errorx.Invalid("外链页面需要有效的 HTTP 或 HTTPS 链接地址")
		}
		menu.Path, menu.Component = "", ""
	default:
		return errorx.Invalid("页面类型无效")
	}
	if menu.MenuName == "" || utf8.RuneCountInString(menu.MenuName) > 50 {
		return errorx.Invalid("菜单名称需为 1 至 50 个字符")
	}
	if menu.MenuType < 1 || menu.MenuType > 3 {
		return errorx.Invalid("菜单类型无效")
	}
	if menu.Status != 0 && menu.Status != 1 {
		return errorx.Invalid("菜单状态无效")
	}
	if menu.MenuType == 3 && menu.Permission == "" {
		return errorx.Invalid("按钮权限标识不能为空")
	}
	if utf8.RuneCountInString(menu.MenuCode) > 100 || utf8.RuneCountInString(menu.Icon) > 100 ||
		utf8.RuneCountInString(menu.Path) > 200 || utf8.RuneCountInString(menu.Component) > 200 ||
		utf8.RuneCountInString(menu.Remark) > 500 {
		return errorx.Invalid("菜单字段长度超过限制")
	}
	return nil
}

func checkMenuUniqueness(repo rbacrepository.SysMenuRepository, menu *rbacentity.SysMenu) error {
	type uniquenessCheck struct {
		value   string
		find    func(string) (*rbacentity.SysMenu, error)
		message string
	}
	checks := []uniquenessCheck{
		{menu.MenuCode, repo.FindByCode, "菜单编码已存在，请使用其他编码"},
		{menu.MenuName, func(name string) (*rbacentity.SysMenu, error) {
			return repo.FindByName(menu.SystemCode, menu.ParentID, name)
		}, "同级菜单名称已存在，请使用其他名称"},
	}
	if menu.MenuType != 3 && menu.Path != "" {
		checks = append(checks, uniquenessCheck{menu.Path, func(path string) (*rbacentity.SysMenu, error) {
			return repo.FindByPath(menu.SystemCode, menu.ParentID, path)
		}, "菜单路径已被使用，请使用其他路径"})
	}
	for _, check := range checks {
		existing, err := check.find(check.value)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return fmt.Errorf("检查菜单唯一性失败: %w", err)
		}
		if existing != nil && existing.ID != menu.ID {
			return errorx.Conflicting(check.message)
		}
	}
	return nil
}

func menuWriteError(err error) error {
	// The unique index remains authoritative when concurrent requests pass
	// the pre-check. Preserve the driver cause without publishing its SQL text.
	if dberror.IsDuplicate(err) {
		return &errorx.PublicError{Kind: errorx.Conflict, Message: "菜单编码已存在，请使用其他编码", Cause: err}
	}
	return err
}

func (s *sysMenuService) Delete(id uint64) error {
	return s.menuRepo.Delete(id)
}

func (s *sysMenuService) GetUserMenus(userID uint64) ([]*rbacentity.SysMenu, error) {
	return s.menuRepo.GetUserMenus(userID)
}

func (s *sysMenuService) GetAllMenus() ([]*rbacentity.SysMenu, error) {
	return s.menuRepo.GetAllMenus()
}
