package manifest

import "testing"

func TestInvalidHostMatchers(t *testing.T) {
	for _, host := range []string{"http://example.org", "site:0", "site:65536", "site:", ":1234", "[site]:443", "*.example.org", "[::1", "a..b", "site/path", " site"} {
		if _, err := ParseHost(host); err == nil {
			t.Errorf("accepted invalid host %q", host)
		}
	}
}
