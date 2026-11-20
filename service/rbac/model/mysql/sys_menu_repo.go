package mysql

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sitehub/common/pkg/logger"
	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

type SysMenuRepositoryImpl struct {
	db *gorm.DB
}

func NewSysMenuRepositoryImpl(db *gorm.DB) rbacrepository.SysMenuRepository {
	return &SysMenuRepositoryImpl{db: db}
}

func (r *SysMenuRepositoryImpl) InTransaction(fn func(rbacrepository.SysMenuRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&SysMenuRepositoryImpl{db: tx})
	})
}

func (r *SysMenuRepositoryImpl) FindByIDForUpdate(id uint64) (*rbacentity.SysMenu, error) {
	var menu rbacentity.SysMenu
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&menu, id).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *SysMenuRepositoryImpl) FindByID(id uint64) (*rbacentity.SysMenu, error) {
	var menu rbacentity.SysMenu
	err := r.db.Preload("Children").First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *SysMenuRepositoryImpl) FindByCode(code string) (*rbacentity.SysMenu, error) {
	var menu rbacentity.SysMenu
	err := r.db.Where("menu_code = ?", code).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *SysMenuRepositoryImpl) FindByName(system string, parentID uint64, name string) (*rbacentity.SysMenu, error) {
	var menu rbacentity.SysMenu
	if name == "" {
		return nil, fmt.Errorf("查询名称为空")
	}
	err := r.db.Where("system_code = ? AND parent_id = ? AND menu_name = ? AND deleted = 0", system, parentID, name).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *SysMenuRepositoryImpl) FindByPath(system string, parentID uint64, path string) (*rbacentity.SysMenu, error) {
	var menu rbacentity.SysMenu
	if path == "" {
		return nil, fmt.Errorf("查询路径为空")
	}
	query := r.db.Where("system_code = ? AND path = ? AND deleted = 0", system, path)
	if !strings.HasPrefix(path, "/") {
		query = query.Where("parent_id = ?", parentID)
	}
	err := query.First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *SysMenuRepositoryImpl) List(page, pageSize int) ([]*rbacentity.SysMenu, int64, error) {
	var menus []*rbacentity.SysMenu
	var total int64

	offset := (page - 1) * pageSize
	err := r.db.Model(&rbacentity.SysMenu{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&menus).Error
	if err != nil {
		return nil, 0, err
	}

	return menus, total, nil
}

func (r *SysMenuRepositoryImpl) GetMenuTree() ([]*rbacentity.SysMenu, error) {
	var menus []*rbacentity.SysMenu
	err := r.db.Where("parent_id = 0").Order("sort ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}

	// 递归加载子菜单
	for _, menu := range menus {
		if err := r.loadChildren(menu); err != nil {
			return nil, err
		}
	}

	return menus, nil
}

func (r *SysMenuRepositoryImpl) GetUserMenus(userID uint64) ([]*rbacentity.SysMenu, error) {
	grants := r.db.Table("sys_role_menu rm").Select("rm.menu_id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Joins("JOIN sys_role r ON r.id = rm.role_id").
		Where("ur.user_id = ? AND r.status = ? AND r.deleted = 0", userID, rbacentity.StatusEnabled)
	active := r.db.Model(&rbacentity.SysMenu{}).Where("status = ? AND deleted = 0", rbacentity.StatusEnabled)
	var batch []*rbacentity.SysMenu
	if err := active.Session(&gorm.Session{}).Where("id IN (?)", grants).Find(&batch).Error; err != nil {
		return nil, err
	}
	var menus []*rbacentity.SysMenu
	seen := make(map[uint64]bool)
	for len(batch) > 0 {
		menus = append(menus, batch...)
		for _, menu := range batch {
			seen[menu.ID] = true
		}
		var parents []uint64
		for _, menu := range batch {
			if menu.ParentID != 0 && !seen[menu.ParentID] {
				seen[menu.ParentID] = true
				parents = append(parents, menu.ParentID)
			}
		}
		if len(parents) == 0 {
			break
		}
		batch = nil
		if err := active.Session(&gorm.Session{}).Where("id IN ?", parents).Find(&batch).Error; err != nil {
			return nil, err
		}
	}
	sort.Slice(menus, func(first, second int) bool {
		if menus[first].Sort == menus[second].Sort {
			return menus[first].ID < menus[second].ID
		}
		return menus[first].Sort < menus[second].Sort
	})
	return r.buildMenuTree(menus, 0), nil
}

func (r *SysMenuRepositoryImpl) Create(menu *rbacentity.SysMenu) error {
	// Map insertion preserves explicit false/zero values despite model defaults.
	now := time.Now()
	menu.CreateTime, menu.UpdateTime = &now, &now
	values := menuEditableValues(menu)
	values["perms"] = menu.Permission
	values["create_by"] = menu.CreateBy
	values["create_time"] = menu.CreateTime
	values["update_by"] = menu.CreateBy
	values["update_time"] = menu.CreateTime
	if err := r.db.Model(&rbacentity.SysMenu{}).Create(values).Error; err != nil {
		return err
	}
	if id, ok := values["id"].(int64); ok {
		menu.ID = uint64(id)
	}
	return nil
}

func (r *SysMenuRepositoryImpl) Update(menu *rbacentity.SysMenu) error {
	// Only editable fields are written. Preserve creation metadata and perms;
	// Save would overwrite them and could insert a missing record.
	values := menuEditableValues(menu)
	values["update_by"] = menu.UpdateBy
	return r.db.Model(&rbacentity.SysMenu{}).Where("id = ?", menu.ID).Updates(values).Error
}

func menuEditableValues(menu *rbacentity.SysMenu) map[string]any {
	return map[string]any{
		"system_code": menu.SystemCode, "perms": menu.Permission,
		"parent_id": menu.ParentID, "menu_name": menu.MenuName,
		"menu_code": menu.MenuCode, "menu_type": menu.MenuType,
		"component": menu.Component, "path": menu.Path, "icon": menu.Icon,
		"page_type": menu.PageType, "external_url": menu.ExternalURL,
		"sort": menu.Sort, "status": menu.Status, "is_visible": menu.IsVisible,
		"remark": menu.Remark,
	}
}

func (r *SysMenuRepositoryImpl) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 递归删除子菜单
		if err := r.deleteChildren(tx, id); err != nil {
			return err
		}
		// 删除菜单
		if err := tx.Delete(&rbacentity.SysMenu{}, id).Error; err != nil {
			return err
		}
		// 删除角色菜单关联
		if err := tx.Where("menu_id = ?", id).Delete(&rbacentity.SysRoleMenu{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *SysMenuRepositoryImpl) GetChildren(parentID uint64) ([]*rbacentity.SysMenu, error) {
	var menus []*rbacentity.SysMenu
	err := r.db.Where("parent_id = ?", parentID).Order("sort ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *SysMenuRepositoryImpl) UpdateStatus(id uint64, status rbacentity.Status) error {
	return r.db.Model(&rbacentity.SysMenu{}).Where("id = ?", id).Update("status", status).Error
}

// loadChildren 递归加载子菜单
func (r *SysMenuRepositoryImpl) loadChildren(menu *rbacentity.SysMenu) error {
	children, err := r.GetChildren(menu.ID)
	if err != nil {
		return err
	}
	menu.Children = children

	for _, child := range children {
		if err := r.loadChildren(child); err != nil {
			return err
		}
	}

	return nil
}

// deleteChildren 递归删除子菜单
func (r *SysMenuRepositoryImpl) deleteChildren(tx *gorm.DB, parentID uint64) error {
	var children []*rbacentity.SysMenu
	if err := tx.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		if err := r.deleteChildren(tx, child.ID); err != nil {
			return err
		}
		if err := tx.Delete(&rbacentity.SysMenu{}, child.ID).Error; err != nil {
			return err
		}
	}

	return nil
}

// buildMenuTree 构建菜单树
func (r *SysMenuRepositoryImpl) buildMenuTree(menus []*rbacentity.SysMenu, parentID uint64) []*rbacentity.SysMenu {
	var tree []*rbacentity.SysMenu
	var menuMap = make(map[uint64]*rbacentity.SysMenu)

	// 构建映射
	for _, menu := range menus {
		menuMap[menu.ID] = menu
	}

	// 构建树
	for _, menu := range menus {
		if menu.ParentID == parentID {
			tree = append(tree, menu)
		} else if parent, exists := menuMap[menu.ParentID]; exists {
			if parent.Children == nil {
				parent.Children = []*rbacentity.SysMenu{}
			}
			parent.Children = append(parent.Children, menu)
		}
	}

	return tree
}

// 获取所有菜单，不带分页
func (r *SysMenuRepositoryImpl) GetAllMenus() (menus []*rbacentity.SysMenu, err error) {
	err = r.db.Find(&menus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("暂无菜单数据", logger.Err("error", err))
			return nil, nil
		}
		return nil, fmt.Errorf("获取菜单出错: %v", err)
	}
	return menus, nil
}
