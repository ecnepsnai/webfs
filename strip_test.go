package webfs

import "testing"

func TestStripURL(t *testing.T) {
	check := func(in, expected string) {
		result := stripURL(in)
		if result != expected {
			t.Errorf("Unexpected stripped URL. Expected '%s' got '%s'", expected, result)
		}
	}

	check("/../../../../../../../etc/passwd", "/etc/passwd")
	check("/~/.ssh/config", "/.ssh/config")
	check("/stupid..filename", "/stupid..filename")
}
