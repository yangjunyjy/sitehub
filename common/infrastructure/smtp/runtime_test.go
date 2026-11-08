package smtp

import (
	"errors"
	"testing"

	"sitehub/common/config"
)

func TestRuntimeSenderUsesNewSnapshotWithoutReplacingInflightSender(t *testing.T) {
	defer SetConfigProvider(nil)
	value := config.EmailConfig{SmtpConfig: config.SmtpConfig{SMTPHost: "first.example.test", Password: "first", PoolSize: 1}}
	SetConfigProvider(func() (config.EmailConfig, error) { return value, nil })
	first, closeFirst, err := acquireSender()
	if err != nil {
		t.Fatal(err)
	}
	defer closeFirst()
	value.SMTPHost = "second.example.test"
	value.Password = "second"
	second, closeSecond, err := acquireSender()
	if err != nil {
		t.Fatal(err)
	}
	defer closeSecond()
	if first.config.SMTPHost != "first.example.test" || second.config.SMTPHost != "second.example.test" || first.config.Password != "first" {
		t.Fatal("in-flight sender snapshot changed")
	}
	if first.codeCfg.HTMLTemplate == "" || first.codeCfg.Length != 6 {
		t.Fatal("verification defaults not applied")
	}
	failure := errors.New("configuration database unavailable")
	SetConfigProvider(func() (config.EmailConfig, error) { return config.EmailConfig{}, failure })
	if _, _, err = acquireSender(); !errors.Is(err, failure) {
		t.Fatalf("provider failure swallowed: %v", err)
	}
}
