package contract

type WebsiteStorage struct {
	Provider    string   `json:"provider"`
	LocalPath   string   `json:"localPath"`
	URLPrefix   string   `json:"urlPrefix"`
	CDN         string   `json:"cdn"`
	Endpoint    string   `json:"endpoint"`
	AccessKey   string   `json:"accessKey"`
	SecretKey   string   `json:"secretKey,omitempty"`
	Bucket      string   `json:"bucket"`
	Region      string   `json:"region"`
	PathStyle   bool     `json:"pathStyle"`
	AllowedExts []string `json:"allowedExts"`
	MaxSize     int64    `json:"maxSize"`
}
