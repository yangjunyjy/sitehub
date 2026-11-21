package contract

// WebsiteContent is the editable content manifest consumed by the product website.
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
	ParticleShape       string   `json:"particleShape"`
	ParticleAutoplay    bool     `json:"particleAutoplay"`
	ParticleInteractive bool     `json:"particleInteractive"`
	Eyebrow             string   `json:"eyebrow"`
	Title               string   `json:"title"`
	Highlight           string   `json:"highlight"`
	Description         string   `json:"description"`
	PrimaryAction       string   `json:"primaryAction"`
	SecondaryAction     string   `json:"secondaryAction"`
	Meta                []string `json:"meta"`
	CapabilityTitle     string   `json:"capabilityTitle"`
	CapabilityIntro     string   `json:"capabilityIntro"`
	FlowTitle           string   `json:"flowTitle"`
	FlowIntro           string   `json:"flowIntro"`
	AudienceTitle       string   `json:"audienceTitle"`
	AudienceIntro       string   `json:"audienceIntro"`
	BlogTitle           string   `json:"blogTitle"`
	CTA                 string   `json:"cta"`
}

type WebsiteCapability struct {
	Icon        string   `json:"icon,omitempty"`
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
