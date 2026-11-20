package schema

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"sitehub/service/rbac/model/entity"
)

type MenuSeed struct {
	Code, Parent, System, Name, Component, Path, Permission string
	Type                                                    entity.MenuType
	Hidden                                                  bool
}

// MenuCatalog is explicit installation data, not a frontend route alias map.
// Components are exact paths under the admin/product/blog application's views.
func MenuCatalog() []MenuSeed {
	var items []MenuSeed
	add := func(code, parent, system, name, component, path, permission string, kind entity.MenuType, hidden bool) {
		items = append(items, MenuSeed{code, parent, system, name, component, path, permission, kind, hidden})
	}
	add("admin-dashboard", "", "admin", "工作台", "dashboard/index", "/dashboard/index", "", 2, false)
	for _, root := range []struct{ code, name string }{{"system", "系统管理"}, {"blog", "博客系统"}, {"product", "产品系统"}, {"admin-settings", "后台管理系统"}} {
		add(root.code, "", "admin", root.name, "", "/dashboard/"+root.code, "", 1, false)
	}
	pages := []struct {
		parent, name, component, permission string
		hidden                              bool
	}{
		{"system", "系统管理", "system/systems/index", "system:system:list", false},
		{"system", "用户管理", "system/users/index", "system:user:list", false},
		{"system", "角色管理", "system/roles/index", "system:role:list", false},
		{"system", "菜单管理", "system/menus/index", "system:menu:list", false},
		{"blog", "文章管理", "blog/articles/index", "blog:article:list", false},
		{"blog", "友链管理", "blog/friends/index", "blog:community:list", false},
		{"blog", "新建文章", "blog/articles/create/index", "blog:article:create", true},
		{"blog", "编辑文章", "blog/articles/editor/index", "blog:article:update", true},
		{"blog", "分类管理", "blog/categories/index", "blog:category:list", false},
		{"blog", "标签管理", "blog/tags/index", "blog:tag:list", false},
		{"blog", "访问统计", "blog/statistics/index", "blog:statistics:list", false},
		{"product", "官网内容", "product/content/index", "product:content:list", false},
		{"product", "下载管理", "product/downloads/index", "product:download:list", false},
		{"product", "下载设置", "product/download-settings/index", "product:download:list", false},
		{"product", "文档管理", "product/docs/index", "product:doc:list", false},
		{"product", "文档模式", "product/docs/modes/index", "product:doc:list", false},
		{"product", "文档分组", "product/docs/groups/index", "product:doc:list", false},
		{"product", "文档编辑", "product/docs/editor/index", "product:doc:edit", true},
		{"product", "文档设置", "product/docs/settings/index", "product:doc:list", false},
		{"product", "文档版本", "product/docs/versions/index", "product:doc:list", false},
		{"admin-settings", "基础设置", "admin/settings/basic/index", "system:settings:list", false},
		{"admin-settings", "存储设置", "admin/settings/storage/index", "file:storage:list", false},
		{"admin-settings", "邮件设置", "admin/settings/email/index", "system:email:manage", false},
	}
	for _, page := range pages {
		code := "page-" + strings.ReplaceAll(page.component, "/", "-")
		part := strings.TrimSuffix(page.component, "/index")
		if page.component == "blog/articles/editor/index" {
			part += "/:id"
		}
		add(code, page.parent, "admin", page.name, page.component, "/dashboard/"+part, page.permission, 2, page.hidden)
	}
	add("admin-profile", "", "admin", "个人中心", "profile/index", "/dashboard/profile", "", 2, true)
	for _, system := range []string{"product", "blog"} {
		add(system+"-home", "", system, "首页", "home/index", "/", "", 2, false)
		if system == "product" {
			add("product-download", "", system, "下载", "download/index", "/download", "", 2, false)
			add("product-docs", "", system, "文档", "docs/index", "/docs", "", 2, false)
		} else {
			for _, page := range []struct{ path, name string }{{"categories", "分类"}, {"tags", "标签"}, {"archives", "归档"}, {"friends", "友链"}, {"about", "关于"}, {"messages", "留言"}} {
				add("blog-"+page.path, "", system, page.name, page.path+"/index", "/"+page.path, "", 2, false)
			}
		}
	}
	// API permissions are separate from menu identifiers. Include existing edit
	// verbs while the feature endpoints transition to explicit create/update verbs.
	permissions := map[string][]string{
		"system":  {"system:system:list", "system:system:create", "system:system:update", "system:system:delete", "system:user:list", "system:user:edit", "system:user:create", "system:user:update", "system:user:delete", "system:role:list", "system:role:edit", "system:role:create", "system:role:update", "system:role:delete", "system:menu:list", "system:menu:edit", "system:menu:create", "system:menu:update", "system:menu:delete", "system:settings:list", "system:settings:edit", "system:email:manage"},
		"blog":    {"blog:article:list", "blog:article:edit", "blog:article:create", "blog:article:update", "blog:article:delete", "blog:article:publish", "blog:article:audit", "blog:category:list", "blog:category:edit", "blog:category:create", "blog:category:update", "blog:category:delete", "blog:tag:list", "blog:tag:edit", "blog:tag:create", "blog:tag:update", "blog:tag:delete", "blog:statistics:list", "blog:settings:list", "blog:settings:edit"},
		"product": {"product:content:list", "product:content:edit", "product:download:list", "product:download:edit", "product:download:create", "product:download:update", "product:download:delete", "product:doc:list", "product:doc:edit", "product:doc:create", "product:doc:update", "product:doc:delete", "product:doc:publish", "file:storage:list", "file:storage:edit"},
	}
	for _, root := range []string{"system", "blog", "product"} {
		if root == "blog" {
			permissions[root] = append(permissions[root], "blog:community:list", "blog:community:review")
		}
		for _, permission := range permissions[root] {
			add("permission-"+permission, root, "admin", permission, "", "", permission, 3, true)
		}
	}
	return items
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
		ids := make(map[string]uint64)
		for index, seed := range MenuCatalog() {
			code := "builtin:" + seed.Code
			var current entity.SysMenu
			err := tx.Where("menu_code = ?", code).Take(&current).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				current = entity.SysMenu{SystemCode: seed.System, ParentID: ids[seed.Parent], MenuCode: code, MenuName: seed.Name, MenuType: seed.Type, Component: seed.Component, Path: seed.Path, Permission: seed.Permission, IsVisible: !seed.Hidden, Status: 1, Sort: index}
				// Map-based create preserves false IsVisible despite GORM defaults.
				if err = tx.Create(&current).Error; err != nil {
					return err
				}
				if seed.Hidden {
					if err = tx.Model(&current).Update("is_visible", false).Error; err != nil {
						return err
					}
				}
			} else if err != nil {
				return err
			}
			ids[seed.Code] = current.ID
		}
		return nil
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
