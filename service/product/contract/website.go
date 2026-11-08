package contract

import "time"

// WebsiteContent is the editable content manifest consumed by the product website.
// It intentionally contains content only; visual presentation remains a frontend concern.
type WebsiteContent struct {
	Brand         WebsiteBrand          `json:"brand"`
	Navigation    []WebsiteNavigation   `json:"navigation"`
	Home          WebsiteHomeContent    `json:"home"`
	Capabilities  []WebsiteCapability   `json:"capabilities"`
	DeliverySteps []WebsiteDeliveryStep `json:"deliverySteps"`
	Audiences     []WebsiteAudience     `json:"audiences"`
	Product       WebsiteProductContent `json:"product"`
	About         WebsiteAboutContent   `json:"about"`
	Footer        WebsiteFooterContent  `json:"footer"`
}

type WebsiteBrand struct {
	Name         string `json:"name"`
	FullName     string `json:"fullName"`
	Description  string `json:"description"`
	ContactEmail string `json:"contactEmail"`
	RecordNumber string `json:"recordNumber"`
	Copyright    string `json:"copyright"`
}

type WebsiteNavigation struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type WebsiteHomeContent struct {
	Eyebrow         string   `json:"eyebrow"`
	Title           string   `json:"title"`
	Highlight       string   `json:"highlight"`
	Description     string   `json:"description"`
	PrimaryAction   string   `json:"primaryAction"`
	SecondaryAction string   `json:"secondaryAction"`
	Meta            []string `json:"meta"`
	CapabilityTitle string   `json:"capabilityTitle"`
	CapabilityIntro string   `json:"capabilityIntro"`
	FlowTitle       string   `json:"flowTitle"`
	FlowIntro       string   `json:"flowIntro"`
	AudienceTitle   string   `json:"audienceTitle"`
	AudienceIntro   string   `json:"audienceIntro"`
	BlogTitle       string   `json:"blogTitle"`
	CTA             string   `json:"cta"`
}

type WebsiteCapability struct {
	Key         string   `json:"key"`
	Index       string   `json:"index"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type WebsiteDeliveryStep struct {
	Index       string `json:"index"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type WebsiteAudience struct {
	Index       string `json:"index"`
	Title       string `json:"title"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
}

type WebsiteProductContent struct {
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	CapabilityTitle    string                     `json:"capabilityTitle"`
	CapabilityIntro    string                     `json:"capabilityIntro"`
	ArchitectureTitle  string                     `json:"architectureTitle"`
	ArchitectureIntro  string                     `json:"architectureIntro"`
	ArchitectureLayers []WebsiteArchitectureLayer `json:"architectureLayers"`
	SecurityTitle      string                     `json:"securityTitle"`
	SecurityIntro      string                     `json:"securityIntro"`
	SecurityItems      []WebsiteSecurityItem      `json:"securityItems"`
}

type WebsiteArchitectureLayer struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type WebsiteSecurityItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type WebsiteAboutContent struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	IntroTitle  string             `json:"introTitle"`
	Paragraphs  []string           `json:"paragraphs"`
	Principles  []WebsitePrinciple `json:"principles"`
	TeamTitle   string             `json:"teamTitle"`
	TeamIntro   string             `json:"teamIntro"`
	TeamRoles   []string           `json:"teamRoles"`
}

type WebsitePrinciple struct {
	Index       string `json:"index"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type WebsiteFooterContent struct {
	Slogan    string `json:"slogan"`
	Statement string `json:"statement"`
}

type WebsiteDownloads struct {
	Title                string                   `json:"title"`
	Groups               []WebsiteDownloadGroup   `json:"groups"`
	Version              string                   `json:"version"`
	ReleaseDate          string                   `json:"releaseDate"`
	SupportArchitectures string                   `json:"supportArchitectures"`
	Description          string                   `json:"description"`
	Packages             []WebsiteDownloadPackage `json:"packages"`
	InstallOptions       []WebsiteInstallOption   `json:"installOptions"`
	Requirements         []WebsiteRequirement     `json:"requirements"`
	SupportTitle         string                   `json:"supportTitle"`
	SupportDescription   string                   `json:"supportDescription"`
}

// Groups are ordered by their position in the download manifest.
type WebsiteDownloadGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WebsiteDownloadPackage struct {
	ID          string `json:"id"`
	GroupID     string `json:"groupId,omitempty"`
	Edition     string `json:"edition,omitempty"`
	SortOrder   int    `json:"sortOrder"`
	Platform    string `json:"platform"`
	Arch        string `json:"arch"`
	Format      string `json:"format"`
	URL         string `json:"url"`
	Checksum    string `json:"checksum"`
	Icon        string `json:"icon"`
	Product     string `json:"product,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Status      string `json:"status,omitempty"`
}

type WebsiteInstallOption struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Badge       string `json:"badge"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

type WebsiteRequirement struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type WebsiteDocs struct {
	Modes          []WebsiteDocMode    `json:"modes"`
	DefaultMode    string              `json:"defaultMode"`
	Revision       uint64              `json:"revision"`
	DefaultVersion string              `json:"defaultVersion"`
	Versions       []WebsiteDocVersion `json:"versions"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
}

type WebsiteDocVersion struct {
	ModeID   string            `json:"modeId"`
	Revision uint64            `json:"revision"`
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Groups   []WebsiteDocGroup `json:"groups"`
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
	Revision       uint64   `json:"revision"`
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
	ID       uint64               `json:"id"`
	Revision uint64               `json:"revision"`
	Title    string               `json:"title"`
	Items    []WebsiteDocDocument `json:"items"`
}

type WebsiteDocDocument struct {
	Revision  uint64    `json:"revision"`
	GroupID   uint64    `json:"groupId"`
	UpdatedAt time.Time `json:"updatedAt"`
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content,omitempty"`
}

type WebsiteDocsNavigation struct {
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
