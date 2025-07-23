package translation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"remnawave-tg-shop-bot/translations"
)

type Translation map[string]string

type Manager struct {
	translations    map[string]Translation
	defaultLanguage string
	mu              sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{
			translations:    make(map[string]Translation),
			defaultLanguage: "en",
		}
	})
	return instance
}

func (tm *Manager) InitFromFS(fsys fs.FS, dir string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	files, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read translation directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(file.Name(), ".json")
		filePath := filepath.Join(dir, file.Name())

		content, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", file.Name(), err)
		}

		var translation Translation
		if err := json.Unmarshal(content, &translation); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", file.Name(), err)
		}

		tm.translations[langCode] = translation
	}

	if _, exists := tm.translations[tm.defaultLanguage]; !exists {
		return fmt.Errorf("default language %s translation not found", tm.defaultLanguage)
	}

	return nil
}

func (tm *Manager) InitTranslations(dir string) error {
	return tm.InitFromFS(os.DirFS(dir), ".")
}

func (tm *Manager) InitDefaultTranslations() error {
	return tm.InitFromFS(translations.FS, ".")
}

func (tm *Manager) GetText(langCode, key string) string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if translation, exists := tm.translations[langCode]; exists {
		if text, exists := translation[key]; exists && text != "" {
			return text
		}
	}

	if translation, exists := tm.translations[tm.defaultLanguage]; exists {
		if text, exists := translation[key]; exists {
			return text
		}
	}

	return key
}
