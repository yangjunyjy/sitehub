package logic

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validate at the RPC boundary so every caller observes the same input contract.
func validateUser(username, nickname, email, password string, state, gender uint64, create bool) error {
	address, err := mail.ParseAddress(email)
	if strings.TrimSpace(username) == "" || utf8.RuneCountInString(username) > 50 || utf8.RuneCountInString(nickname) > 50 ||
		err != nil || address.Address != email || len(email) > 100 || state > 1 || gender > 2 {
		return status.Error(codes.InvalidArgument, "invalid user fields")
	}
	if create || password != "" {
		return validatePassword(password)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return status.Error(codes.InvalidArgument, "password must contain 8–72 bytes")
	}
	return nil
}

func validateRole(code, name, description string, state int64) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || utf8.RuneCountInString(code) > 50 ||
		utf8.RuneCountInString(name) > 50 || utf8.RuneCountInString(description) > 200 || (state != 0 && state != 1) {
		return status.Error(codes.InvalidArgument, "invalid role fields")
	}
	return nil
}
