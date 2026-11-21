package schema

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gorm.io/gorm"
	"sitehub/service/rbac/model/entity"
)

type MenuSeed struct {
	Code, Parent, System, Name, Component, Path, Permission string
	ExternalURL                                             string
	Type                                                    entity.MenuType
	Hidden                                                  bool
}

// MenuCatalog discovers pages and navigation directories from the shipped applications.
func MenuCatalog() ([]MenuSeed, error) {
	items, err := discoverMenus()
	if err != nil {
		return nil, err
	}
	// API permissions are separate from menu identifiers. Include existing edit
	// verbs while the feature endpoints transition to explicit create/update verbs.
	permissions := map[string][]string{
		"system":  {"system:system:list", "system:system:create", "system:system:update", "system:system:delete", "system:user:list", "system:user:edit", "system:user:create", "system:user:update", "system:user:delete", "system:role:list", "system:role:edit", "system:role:create", "system:role:update", "system:role:delete", "system:menu:list", "system:menu:edit", "system:menu:create", "system:menu:update", "system:menu:delete", "system:icon:list", "system:icon:create", "system:icon:update", "system:icon:delete", "system:settings:list", "system:settings:edit", "system:email:manage"},
		"blog":    {"blog:article:list", "blog:article:edit", "blog:article:create", "blog:article:update", "blog:article:delete", "blog:article:publish", "blog:article:audit", "blog:category:list", "blog:category:edit", "blog:category:create", "blog:category:update", "blog:category:delete", "blog:tag:list", "blog:tag:edit", "blog:tag:create", "blog:tag:update", "blog:tag:delete", "blog:statistics:list", "blog:settings:list", "blog:settings:edit"},
		"product": {"website:content:edit", "website:docs:edit", "website:downloads:edit", "product:content:list", "product:content:edit", "product:download:list", "product:download:edit", "product:download:create", "product:download:update", "product:download:delete", "product:doc:list", "product:doc:edit", "product:doc:create", "product:doc:update", "product:doc:delete", "product:doc:publish", "file:storage:list", "file:storage:edit"},
	}
	permissions["system"] = append(permissions["system"], "gateway:config:read", "gateway:config:publish", "gateway:config:rollback")
	if address := strings.TrimSpace(os.Getenv("GATEWAY_CONSOLE_URL")); address != "" {
		u, err := url.Parse(address)
		if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("GATEWAY_CONSOLE_URL must be an HTTP(S) console address")
		}
		items = append(items, MenuSeed{Code: "gateway-console", Parent: "system", System: "admin", Name: "网关控制台", ExternalURL: address, Type: 2})
	}
	for _, root := range []string{"system", "blog", "product"} {
		if root == "blog" {
			permissions[root] = append(permissions[root], "blog:community:list", "blog:community:review")
		}
		for _, permission := range permissions[root] {
			items = append(items, MenuSeed{Code: "permission-" + permission, Parent: root, System: "admin", Name: permission, Permission: permission, Type: 3, Hidden: true})
		}
	}
	return items, nil
}

// SeedCatalog runs only through explicit bootstrap. Existing stable codes and
// customized values are preserved. Repeating the command creates no duplicates.
func SeedCatalog(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Preserve legacy button grants when moving their previously overloaded
		// identifier into the independent permission column; never replace a grant.
		if err := tx.Model(&entity.SysMenu{}).Where("menu_type = ? AND (perms IS NULL OR perms = ?) AND menu_code <> ?", 3, "", "").Update("perms", gorm.Expr("menu_code")).Error; err != nil {
			return err
		}
		for _, value := range []entity.SysSystem{{Code: "admin", Name: "后台管理系统", Enabled: true}, {Code: "product", Name: "产品系统", Enabled: true, Public: true}, {Code: "blog", Name: "博客系统", Enabled: true, Public: true}} {
			var current entity.SysSystem
			if err := tx.Where("code = ?", value.Code).Take(&current).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&value).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		if err := tx.Model(&entity.SysMenu{}).Where("menu_code = ? AND component = ?", "builtin:product-docs", "docs/[[id]]/index").Update("component", "docs/index").Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.SysMenu{}).Where("menu_code = ?", "builtin:page-blog-articles-editor-[id]-index").Update("menu_code", "builtin:page-blog-articles-editor-index").Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.SysMenu{}).Where("component = ? AND system_code = ?", "blog/articles/editor/[id]/index", "admin").Update("component", "blog/articles/editor/index").Error; err != nil {
			return err
		}
		var previousFriendsMenu uint64
		ids := make(map[string]uint64)
		seeds, err := MenuCatalog()
		if err != nil {
			return err
		}
		for index, seed := range seeds {
			code := "builtin:" + seed.Code
			var current entity.SysMenu
			err := tx.Where("menu_code = ?", code).Take(&current).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				current = entity.SysMenu{SystemCode: seed.System, ParentID: ids[seed.Parent], MenuCode: code, MenuName: seed.Name, MenuType: seed.Type, Component: seed.Component, Path: seed.Path, Permission: seed.Permission, IsVisible: !seed.Hidden, Status: 1, Sort: index}
				if seed.ExternalURL != "" {
					current.PageType = "external"
					current.ExternalURL = seed.ExternalURL
				}
				// Map-based create preserves false IsVisible despite GORM defaults.
				if err = tx.Create(&current).Error; err != nil {
					return err
				}
				if previousFriendsMenu != 0 && (seed.Component == "blog/messages/index" || seed.Component == "blog/comments/index") {
					// Preserve access to the two queues split from the original friends page.
					if err = tx.Exec("INSERT INTO sys_role_menu (role_id, menu_id, created_at) SELECT DISTINCT role_id, ?, CURRENT_TIMESTAMP FROM sys_role_menu WHERE menu_id = ?", current.ID, previousFriendsMenu).Error; err != nil {
						return err
					}
				}
				if seed.Hidden {
					if err = tx.Model(&current).Update("is_visible", false).Error; err != nil {
						return err
					}
				}
			} else if err != nil {
				return err
			} else if seed.Component == "blog/friends/index" {
				previousFriendsMenu = current.ID
			}
			ids[seed.Code] = current.ID
		}
		return retireProductPages(tx)
	})
}

func AssociateAdminMenus(db *gorm.DB) error {
	var role entity.SysRole
	if err := db.Where("id = ? AND role_code = ? AND status = ? AND deleted = ?", entity.SuperAdminRoleID, "super_admin", 1, 0).Take(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		return err
	}
	var menus []entity.SysMenu
	if err := db.Where("system_code = ? AND deleted = ?", "admin", 0).Find(&menus).Error; err != nil {
		return err
	}
	for _, menu := range menus {
		var count int64
		if err := db.Model(&entity.SysRoleMenu{}).Where("role_id = ? AND menu_id = ?", role.ID, menu.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&entity.SysRoleMenu{RoleID: role.ID, MenuID: menu.ID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Remove the built-in pages that moved into the two product workspaces.
// Promote custom children one level within their own tree, keeping their grants.
func retireProductPages(tx *gorm.DB) error {
	for _, path := range []string{"product/docs/groups", "product/docs/modes", "product/docs/settings", "product/docs/versions", "product/download-settings"} {
		component := path + "/index"
		var retired entity.SysMenu
		err := tx.Where("menu_code = ? AND component = ? AND system_code = ? AND deleted = 0", "builtin:page-"+strings.ReplaceAll(component, "/", "-"), component, "admin").Take(&retired).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if err = tx.Model(&entity.SysMenu{}).Where("parent_id = ?", retired.ID).Update("parent_id", retired.ParentID).Error; err != nil {
			return err
		}
		if err = tx.Where("menu_id = ?", retired.ID).Delete(&entity.SysRoleMenu{}).Error; err != nil {
			return err
		}
		if err = tx.Model(&retired).Updates(map[string]any{"deleted": 1, "is_visible": false, "status": 0}).Error; err != nil {
			return err
		}
	}
	return nil
}
