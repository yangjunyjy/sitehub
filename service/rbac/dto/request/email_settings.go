package request

// UpdateEmailSettingsRequest changes SMTP settings; an omitted password preserves it.
type UpdateEmailSettingsRequest struct {
	SMTPHost  string  `json:"smtpHost"`
	SMTPPort  int     `json:"smtpPort"`
	FromEmail string  `json:"fromEmail"`
	FromName  string  `json:"fromName"`
	Subject   string  `json:"subject"`
	Password  *string `json:"password,omitempty"`
}
