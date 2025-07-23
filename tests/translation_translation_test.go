package tests

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func TestLoadTranslationsFallback(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	if got := tm.GetText("en", "account_button"); got == "account_button" {
		t.Fatalf("expected translation for account_button")
	}
	en := tm.GetText("es", "account_button")
	if en != tm.GetText("en", "account_button") {
		t.Fatalf("fallback not used")
	}
}

func loadYAML(path string) (map[string]string, error) {
	//nolint:gosec // reading test fixture
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func collectUsedKeys(t *testing.T) map[string]struct{} {
	keys := make(map[string]struct{})
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || strings.HasPrefix(path, ".git") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, string(filepath.Separator)+"tests"+string(filepath.Separator)) {
			return nil
		}
		//nolint:gosec // reading source files for keys
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		re := regexp.MustCompile(`GetText\([^,]+,\s*"([^"]+)"`)
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			keys[m[1]] = struct{}{}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk dir: %v", err)
	}
	return keys
}

func TestTranslationsConsistency(t *testing.T) {
	enMap, err := loadYAML("../translations/en.yml")
	if err != nil {
		t.Fatalf("load en: %v", err)
	}
	ruMap, err := loadYAML("../translations/ru.yml")
	if err != nil {
		t.Fatalf("load ru: %v", err)
	}
	if len(enMap) != len(ruMap) {
		t.Fatalf("key count mismatch: en=%d ru=%d", len(enMap), len(ruMap))
	}
	for k := range enMap {
		if _, ok := ruMap[k]; !ok {
			t.Errorf("missing key in ru: %s", k)
		}
	}
	used := collectUsedKeys(t)
	for k := range used {
		if _, ok := enMap[k]; !ok {
			t.Errorf("key used in code but missing in translations: %s", k)
		}
	}
	for k := range enMap {
		if _, ok := used[k]; !ok {
			t.Errorf("unused translation key: %s", k)
		}
	}
}
