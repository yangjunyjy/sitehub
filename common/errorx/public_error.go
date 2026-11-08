package errorx

// ErrorKind describes a business failure without coupling services to HTTP.
type ErrorKind uint8

const (
	InvalidInput ErrorKind = iota + 1
	Conflict
	NotFound
)

// PublicError contains an explicitly safe message. Cause is retained for diagnostics
// and errors.Is/As; it must never be appended to the client-facing message.
type PublicError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func (e *PublicError) Error() string { return e.Message }
func (e *PublicError) Unwrap() error { return e.Cause }

func Invalid(message string) error     { return &PublicError{Kind: InvalidInput, Message: message} }
func Conflicting(message string) error { return &PublicError{Kind: Conflict, Message: message} }
