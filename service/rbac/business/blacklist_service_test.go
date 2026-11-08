package services

import (
	"strings"
	"testing"
)

func TestBlacklistNotificationUsesInjectedSender(t *testing.T) {
	var sent int
	send := func(to, subject, html string) error {
		sent++
		if to != "alerts@example.test" || !strings.Contains(subject, "192.0.2.1") || !strings.Contains(html, "test reason") {
			t.Fatal("notification lost configured recipient or event details")
		}
		return nil
	}
	service := NewBlacklistService(nil, send, " alerts@example.test ")
	service.sendNotificationEmail("192.0.2.1", "test reason")
	if sent != 1 {
		t.Fatal("injected sender was not called")
	}
	NewBlacklistService(nil, send, "  ").sendNotificationEmail("192.0.2.1", "test reason")
	if sent != 1 {
		t.Fatal("empty recipient should disable notifications")
	}
}
