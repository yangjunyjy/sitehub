package services

import (
	"strings"

	productcontract "sitehub/service/product/contract"
	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

func modeView(m productdomain.DocMode) productcontract.WebsiteDocMode {
	return productcontract.WebsiteDocMode{ID: m.ID, Label: m.Label, Description: m.Description, Icon: m.Icon,
		Status: m.Status, Available: m.Status == "published", DefaultVersion: m.DefaultVersion,
		SortOrder: m.SortOrder, Revision: m.Revision}
}
func findMode(r productrepository.ProductDocsRepository, id string) (productdomain.DocMode, error) {
	modes, err := r.Modes()
	if err != nil {
		return productdomain.DocMode{}, err
	}
	for _, m := range modes {
		if m.ID == id {
			return m, nil
		}
	}
	return productdomain.DocMode{}, ErrWebsiteDocumentNotFound
}
func defaultMode(versions []productdomain.DocVersion, version string) string {
	for _, v := range versions {
		if v.ID == version {
			return v.ModeID
		}
	}
	return ""
}

func (s *websiteContentService) GetDocModes() (result productcontract.WebsiteDocModes, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		site, e := r.Site(false)
		if e != nil {
			return e
		}
		modes, e := r.Modes()
		if e != nil {
			return e
		}
		versions, e := r.Versions()
		if e != nil {
			return e
		}
		result = productcontract.WebsiteDocModes{Title: site.Title, Description: site.Description,
			DefaultMode: defaultMode(versions, site.DefaultVersion), Modes: []productcontract.WebsiteDocMode{}}
		for _, m := range modes {
			if m.Status != "hidden" {
				entry := modeView(m)
				// Expose only published versions owned by this available mode.
				if entry.Available {
					for _, v := range versions {
						if v.ModeID == m.ID && v.Status == "published" {
							entry.VersionIDs = append(entry.VersionIDs, v.ID)
						}
					}
				}
				result.Modes = append(result.Modes, entry)
			}
		}
		return nil
	})
	return
}

// 与版本创建使用同一文档站锁：引用检查与删除在同一事务中完成。
func (s *websiteContentService) ChangeDocMode(in productdomain.DocModeInput) error {
	if !docVersionID.MatchString(in.ID) {
		return invalidDoc("模式标识无效")
	}
	return s.mutateDocs(func(r productrepository.ProductDocsRepository, site *productdomain.DocSite) error {
		modes, err := r.Modes()
		if err != nil {
			return err
		}
		versions, err := r.Versions()
		if err != nil {
			return err
		}
		var current *productdomain.DocMode
		for i := range modes {
			if modes[i].ID == in.ID {
				current = &modes[i]
			}
		}
		if current == nil {
			if in.Action != "save" || in.Revision != 0 {
				return ErrWebsiteDocumentNotFound
			}
			if len(modes) >= 50 {
				return invalidDoc("最多保留 50 个文档模式")
			}
		} else if current.Revision != in.Revision {
			return ErrWebsiteDocsConflict
		}
		if in.Action == "delete" {
			for _, v := range versions {
				if v.ModeID == in.ID {
					return invalidDoc("该模式仍有关联版本，请先处理所属版本后再删除")
				}
			}
			return r.DeleteMode(in.ID)
		}
		if in.Action == "default" {
			if current.Status != "published" || current.DefaultVersion == "" {
				return invalidDoc("默认模式必须已上线且有默认版本")
			}
			site.DefaultVersion = current.DefaultVersion
			site.Revision++
			return r.PutSite(*site)
		}
		if in.Action != "save" {
			return invalidDoc("模式操作无效")
		}
		in.Label = strings.TrimSpace(in.Label)
		if in.Label == "" || len([]rune(in.Label)) > 60 || len([]rune(in.Description)) > 500 || in.SortOrder < 0 || in.SortOrder > 9999 {
			return invalidDoc("请填写有效的模式名称、说明和排序")
		}
		if in.Status != "published" && in.Status != "upcoming" && in.Status != "hidden" {
			return invalidDoc("模式状态无效")
		}
		switch in.Icon {
		case "terminal", "assets", "pipeline", "playbook", "artifact", "shield":
		default:
			return invalidDoc("模式图标无效")
		}
		for _, m := range modes {
			if m.ID != in.ID && strings.EqualFold(m.Label, in.Label) {
				return invalidDoc("模式名称已存在")
			}
		}
		if in.DefaultVersion != "" {
			valid := false
			for _, v := range versions {
				if v.ID == in.DefaultVersion && v.ModeID == in.ID && v.Status == "published" {
					valid = true
				}
			}
			if !valid {
				return invalidDoc("默认版本必须是当前模式下已发布的版本")
			}
		}
		if in.Status == "published" && in.DefaultVersion == "" {
			return invalidDoc("上线前请先创建、发布所属版本，并选择默认版本")
		}
		isDefault := defaultMode(versions, site.DefaultVersion) == in.ID
		if isDefault && in.Status != "published" {
			return invalidDoc("请先切换默认模式，再隐藏或下线当前模式")
		}
		revision := uint64(1)
		if current != nil {
			revision = current.Revision + 1
		}
		if err = r.PutMode(productdomain.DocMode{ID: in.ID, Label: in.Label, Description: strings.TrimSpace(in.Description),
			Icon: in.Icon, Status: in.Status, DefaultVersion: in.DefaultVersion, SortOrder: in.SortOrder, Revision: revision}); err != nil {
			return err
		}
		if isDefault && site.DefaultVersion != in.DefaultVersion {
			site.DefaultVersion = in.DefaultVersion
			site.Revision++
			return r.PutSite(*site)
		}
		return nil
	})
}

// 即使绕过前台直接请求正文，也不能跨模式读取或读取隐藏/待推出模式。
func publicDocVersion(r productrepository.ProductDocsRepository, site productdomain.DocSite, version, mode string) (productdomain.DocVersion, error) {
	if mode == "" {
		target := version
		if target == "" {
			target = site.DefaultVersion
		}
		v, err := findVersion(r, target)
		if err != nil {
			return v, err
		}
		mode = v.ModeID
	}
	m, err := findMode(r, mode)
	if err != nil {
		return productdomain.DocVersion{}, err
	}
	if m.Status != "published" {
		return productdomain.DocVersion{}, ErrWebsiteDocumentNotFound
	}
	if version == "" {
		version = m.DefaultVersion
	}
	v, err := findVersion(r, version)
	if err != nil {
		return v, err
	}
	if v.ModeID != m.ID || v.Status != "published" {
		return productdomain.DocVersion{}, ErrWebsiteDocumentNotFound
	}
	return v, nil
}
