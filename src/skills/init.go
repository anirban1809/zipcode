package skills

import (
	"os"
	"path/filepath"
	"strings"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func ensureDir(path string) {
	_ = os.MkdirAll(path, 0755)
}

func Init(internalDir, globalDir, projectDir, stateFile string) (*SkillRegistry, *Watcher, error) {
	internalDir = expandHome(internalDir)
	globalDir = expandHome(globalDir)
	projectDir = expandHome(projectDir)
	stateFile = expandHome(stateFile)

	ensureDir(globalDir)
	ensureDir(projectDir)

	registry := NewSkillRegistry()
	if err := registry.SetStateFile(stateFile); err != nil {
		return nil, nil, err
	}

	merged, err := LoadAll(internalDir, globalDir, projectDir)
	if err != nil {
		return nil, nil, err
	}
	registry.replace(merged)

	watcher, err := StartWatcher(registry, internalDir, globalDir, projectDir)
	if err != nil {
		return registry, nil, err
	}

	return registry, watcher, nil
}
