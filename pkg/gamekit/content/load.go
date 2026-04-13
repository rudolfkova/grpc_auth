package content

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Bundle каталог + загруженные сценарии + общий Runner (движок дописывает свои op).
type Bundle struct {
	Catalog   *Catalog
	Scenarios map[string]*Scenario
	Runner    *Runner
}

// LoadCatalog читает только catalog.json.
func LoadCatalog(path string) (*Catalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("content catalog read %q: %w", path, err)
	}
	var c Catalog
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("content catalog json %q: %w", path, err)
	}
	if c.Items == nil {
		c.Items = make(map[string]ItemDef)
	}
	return &c, nil
}

// loadScenariosDir читает все *.json из каталога; ключ — имя файла (например lever.json).
func loadScenariosDir(dir string) (map[string]*Scenario, error) {
	out := make(map[string]*Scenario)
	if strings.TrimSpace(dir) == "" {
		return out, nil
	}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".json") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var sc Scenario
		if err := json.Unmarshal(raw, &sc); err != nil {
			return fmt.Errorf("scenario %q: %w", path, err)
		}
		key := filepath.Base(path)
		out[key] = &sc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LoadBundle загружает каталог и сценарии, проверяет что у каждого interact есть файл.
func LoadBundle(catalogPath, scriptsDir string) (*Bundle, error) {
	cat, err := LoadCatalog(catalogPath)
	if err != nil {
		return nil, err
	}
	scenarios, err := loadScenariosDir(scriptsDir)
	if err != nil {
		return nil, err
	}
	for id, def := range cat.Items {
		if def.Interact == nil {
			continue
		}
		script := strings.TrimSpace(def.Interact.Script)
		if script == "" {
			return nil, fmt.Errorf("content: item %q has empty interact.script", id)
		}
		if _, ok := scenarios[script]; !ok {
			return nil, fmt.Errorf("content: item %q references script %q not found under %q", id, script, scriptsDir)
		}
	}
	return &Bundle{
		Catalog:   cat,
		Scenarios: scenarios,
		Runner:    NewRunner(),
	}, nil
}
