package services

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sitehub/common/errorx"
	"sitehub/service/rbac/model/entity"
)

var systemCodePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,49}$`)

type SysSystemService struct{ db *gorm.DB }

func NewSysSystemService(db *gorm.DB) *SysSystemService { return &SysSystemService{db: db} }

func (s *SysSystemService) List(ctx context.Context) ([]*entity.SysSystem, error) {
	var items []*entity.SysSystem
	err := s.db.WithContext(ctx).Where("deleted = ?", 0).Order("id asc").Find(&items).Error
	return items, err
}

func validateSystem(value *entity.SysSystem) error {
	value.Code, value.Name, value.Origin = strings.TrimSpace(value.Code), strings.TrimSpace(value.Name), strings.TrimSpace(value.Origin)
	if !systemCodePattern.MatchString(value.Code) || value.Name == "" || utf8.RuneCountInString(value.Name) > 100 || utf8.RuneCountInString(value.Description) > 500 {
		return errorx.Invalid("系统代码、名称或描述无效")
	}
	if value.Code == "admin" && value.Public {
		return errorx.Invalid("后台管理系统不能开放公共菜单")
	}
	if value.Origin != "" {
		u, err := url.Parse(value.Origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return errorx.Invalid("系统地址必须是 HTTP/HTTPS Origin")
		}
		value.Origin = u.Scheme + "://" + u.Host
	}
	return nil
}

func (s *SysSystemService) Save(ctx context.Context, value *entity.SysSystem) error {
	if err := validateSystem(value); err != nil {
		return err
	}
	return menuWriteError(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if value.ID == 0 {
			return tx.Create(value).Error
		}
		var current entity.SysSystem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted = ?", value.ID, 0).Take(&current).Error; err != nil {
			return err
		}
		if current.Code != value.Code {
			return errorx.Invalid("已有系统代码不可修改")
		}
		return tx.Model(&current).Updates(map[string]any{"name": value.Name, "description": value.Description, "origin": value.Origin, "enabled": value.Enabled, "public": value.Public}).Error
	}))
}

func (s *SysSystemService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return errorx.Invalid("系统ID无效")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.SysSystem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted = ?", id, 0).Take(&current).Error; err != nil {
			return err
		}
		if current.Code == "admin" {
			return errorx.Invalid("不能删除后台管理系统")
		}
		var count int64
		if err := tx.Model(&entity.SysMenu{}).Where("system_code = ? AND deleted = ?", current.Code, 0).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errorx.Conflicting("请先删除该系统下的菜单")
		}
		return tx.Delete(&current).Error
	})
}

// ValidateMenu enforces that a parent and child belong to the same system.
func (s *SysSystemService) ValidateMenu(ctx context.Context, value *entity.SysMenu) error {
	if value.SystemCode == "" {
		value.SystemCode = "admin"
	}
	var system entity.SysSystem
	if err := s.db.WithContext(ctx).Where("code = ? AND deleted = ?", value.SystemCode, 0).Take(&system).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.Invalid("所属系统不存在")
		}
		return err
	}
	if value.ParentID == value.ID && value.ID != 0 {
		return errorx.Invalid("菜单不能以自身为父级")
	}
	if value.ParentID != 0 {
		var parent entity.SysMenu
		if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", value.ParentID, 0).Take(&parent).Error; err != nil {
			return err
		}
		if parent.SystemCode != value.SystemCode {
			return errorx.Invalid("父菜单必须属于同一系统")
		}
	}
	if value.ID != 0 {
		var current entity.SysMenu
		if err := s.db.WithContext(ctx).Where("id = ?", value.ID).Take(&current).Error; err != nil {
			return err
		}
		if current.SystemCode != value.SystemCode {
			var children int64
			if err := s.db.WithContext(ctx).Model(&entity.SysMenu{}).Where("parent_id = ? AND deleted = ?", value.ID, 0).Count(&children).Error; err != nil {
				return err
			}
			if children > 0 {
				return errorx.Invalid("请先迁移子菜单，再更改所属系统")
			}
		}
	}
	return nil
}

func (s *SysSystemService) MenuTree(ctx context.Context, code string, public bool) ([]*entity.SysMenu, error) {
	if code == "" {
		code = "admin"
	}
	if !systemCodePattern.MatchString(code) {
		return nil, errorx.Invalid("系统代码无效")
	}
	var system entity.SysSystem
	if err := s.db.WithContext(ctx).Where("code = ? AND deleted = ?", code, 0).Take(&system).Error; err != nil {
		return nil, err
	}
	if public && (code == "admin" || !system.Enabled || !system.Public) {
		return []*entity.SysMenu{}, nil
	}
	var items []*entity.SysMenu
	query := s.db.WithContext(ctx).Where("system_code = ? AND deleted = ?", code, 0)
	if public {
		query = query.Where("status = ? AND is_visible = ? AND menu_type IN ?", 1, true, []int{1, 2})
	}
	if err := query.Order("sort asc,id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint64]*entity.SysMenu, len(items))
	for _, item := range items {
		item.Children = nil
		byID[item.ID] = item
	}
	var roots []*entity.SysMenu
	for _, item := range items {
		if item.ParentID == 0 {
			roots = append(roots, item)
		} else if parent := byID[item.ParentID]; parent != nil {
			parent.Children = append(parent.Children, item)
		}
	}
	return roots, nil
}

func AdminMenuTree(items []*entity.SysMenu) []*entity.SysMenu {
	var result []*entity.SysMenu
	for _, item := range items {
		if item != nil && (item.SystemCode == "admin" || item.SystemCode == "") {
			copy := *item
			copy.Children = AdminMenuTree(item.Children)
			result = append(result, &copy)
		}
	}
	return result
}
