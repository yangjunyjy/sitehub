package contract

import "time"

type ManualWorkspace struct {
	ModeID         string              `json:"modeId"`
	VersionID      string              `json:"versionId"`
	Groups         []WebsiteDocGroup   `json:"groups"`
	Modes          []WebsiteDocMode    `json:"modes"`
	DefaultMode    string              `json:"defaultMode"`
	Revision       uint64              `json:"revision,string"`
	DefaultVersion string              `json:"defaultVersion"`
	Versions       []WebsiteDocVersion `json:"versions"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
}

type WebsiteDocVersion struct {
	ModeID   string `json:"modeId"`
	Revision uint64 `json:"revision,string"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}
type WebsiteDocMode struct {
	VersionIDs     []string `json:"versionIds,omitempty"`
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	Description    string   `json:"description"`
	Icon           string   `json:"icon"`
	Status         string   `json:"status"`
	Available      bool     `json:"available"`
	DefaultVersion string   `json:"defaultVersion"`
	SortOrder      int      `json:"sortOrder"`
	Revision       uint64   `json:"revision,string"`
}
type WebsiteDocModes struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	DefaultMode string           `json:"defaultMode"`
	Modes       []WebsiteDocMode `json:"modes"`
}

type WebsiteDocVersionSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebsiteDocGroup struct {
	ID       uint64                  `json:"id,string"`
	Revision uint64                  `json:"revision,string"`
	Title    string                  `json:"title"`
	Items    []ManualDocumentSummary `json:"items"`
}

type ManualDocumentSummary struct {
	PublishedRevision uint64    `json:"publishedRevision,string"`
	Status            string    `json:"status"`
	Revision          uint64    `json:"revision,string"`
	GroupID           uint64    `json:"groupId,string"`
	UpdatedAt         time.Time `json:"updatedAt"`
	ID                string    `json:"id"`
	Slug              string    `json:"slug"`
	Title             string    `json:"title"`
	Summary           string    `json:"summary"`
}

type WebsiteDocDocument struct {
	ManualDocumentSummary
	Content string `json:"content"`
}

type WebsiteDocCatalog struct {
	WebsiteDocModes
	ModeID     string                 `json:"modeId"`
	Navigation *WebsiteDocsNavigation `json:"navigation"`
}

type WebsiteDocsNavigation struct {
	ModeID         string                      `json:"modeId"`
	VersionID      string                      `json:"versionId"`
	DefaultVersion string                      `json:"defaultVersion"`
	Versions       []WebsiteDocVersionSummary  `json:"versions"`
	Title          string                      `json:"title"`
	Description    string                      `json:"description"`
	Version        string                      `json:"version"`
	Groups         []WebsiteDocNavigationGroup `json:"groups"`
}

type WebsiteDocNavigationGroup struct {
	Title string                         `json:"title"`
	Items []WebsiteDocNavigationDocument `json:"items"`
}

type WebsiteDocNavigationDocument struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
