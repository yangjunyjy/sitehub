package services

import (
	productcontract "sitehub/service/product/contract"
	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

// The reader sees modes, published versions and the selected sidebar from one snapshot.
// Document bodies remain individually addressable and are never loaded for a catalog.
func (s *ManualService) GetDocCatalog(version, mode string) (result productcontract.WebsiteDocCatalog, err error) {
	err = s.docs.Transaction(func(r productrepository.ProductDocsRepository) error {
		modes, versions, err := readDocModes(r)
		if err != nil {
			return err
		}
		result.WebsiteDocModes = modes
		if mode == "" && version != "" {
			mode = defaultMode(versions, version)
		}
		if mode == "" {
			mode = modes.DefaultMode
			if mode == "" && version == "" && len(modes.Modes) > 0 {
				mode = modes.Modes[0].ID
			}
		}
		for _, entry := range modes.Modes {
			if entry.ID != mode {
				continue
			}
			result.ModeID = mode
			if !entry.Available {
				if version != "" {
					return ErrWebsiteDocumentNotFound
				}
				return nil
			}
			if version == "" {
				version = entry.DefaultVersion
			}
			navigation, err := publishedNavigation(r, modes, versions, entry, version)
			if err != nil {
				return err
			}
			result.Navigation = &navigation
			return nil
		}
		if mode == "" && len(modes.Modes) == 0 && version == "" {
			return nil
		}
		return ErrWebsiteDocumentNotFound
	})
	return
}

func publishedNavigation(r productrepository.ProductDocsRepository, site productcontract.WebsiteDocModes, versions []productdomain.DocVersion, mode productcontract.WebsiteDocMode, version string) (productcontract.WebsiteDocsNavigation, error) {
	nav := productcontract.WebsiteDocsNavigation{Title: site.Title, Description: site.Description, ModeID: mode.ID, DefaultVersion: mode.DefaultVersion,
		Versions: []productcontract.WebsiteDocVersionSummary{}, Groups: []productcontract.WebsiteDocNavigationGroup{}}
	for _, v := range versions {
		if v.ModeID != mode.ID || v.Status != "published" {
			continue
		}
		nav.Versions = append(nav.Versions, productcontract.WebsiteDocVersionSummary{ID: v.ID, Name: v.Name})
		if v.ID == version {
			nav.VersionID, nav.Version = v.ID, v.Name
		}
	}
	if nav.VersionID == "" {
		return nav, ErrWebsiteDocumentNotFound
	}
	groups, err := r.Groups(version)
	if err != nil {
		return nav, err
	}
	docs, err := r.PublishedDocuments(version)
	if err != nil {
		return nav, err
	}
	byGroup := make(map[uint64][]productcontract.WebsiteDocNavigationDocument)
	for _, d := range docs {
		byGroup[d.GroupID] = append(byGroup[d.GroupID], productcontract.WebsiteDocNavigationDocument{ID: d.DocKey, Slug: d.Slug, Title: d.Title, Summary: d.Summary})
	}
	for _, g := range groups {
		items := byGroup[g.ID]
		if items == nil {
			items = []productcontract.WebsiteDocNavigationDocument{}
		}
		nav.Groups = append(nav.Groups, productcontract.WebsiteDocNavigationGroup{Title: g.Title, Items: items})
	}
	return nav, nil
}
