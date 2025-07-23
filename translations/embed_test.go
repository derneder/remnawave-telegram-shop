package translations

import "testing"

func TestEmbeddedFiles(t *testing.T) {
	if _, err := FS.ReadFile("en.yml"); err != nil {
		t.Fatalf("failed to read en.yml: %v", err)
	}
}
