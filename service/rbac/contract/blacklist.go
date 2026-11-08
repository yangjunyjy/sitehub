package contract

// Blacklist exposes the capabilities required by request rate limiting.
// Persistence entities and repositories stay within rbac.
type Blacklist interface {
	BlockedIPs() ([]string, error)
	CheckIPBlocked(ip string) (bool, error)
	AddToBlacklist(ip, reason string) error
	RemoveFromBlacklist(id uint) error
}
