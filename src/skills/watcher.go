package skills

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	internalDir string
	globalDir   string
	projectDir  string
	registry    *SkillRegistry
	fsw         *fsnotify.Watcher
	stopOnce    sync.Once
	stopCh      chan struct{}
}

func StartWatcher(registry *SkillRegistry, internalDir, globalDir, projectDir string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		internalDir: internalDir,
		globalDir:   globalDir,
		projectDir:  projectDir,
		registry:    registry,
		fsw:         fsw,
		stopCh:      make(chan struct{}),
	}

	for _, dir := range []string{internalDir, globalDir, projectDir} {
		if dir == "" {
			continue
		}
		_ = fsw.Add(dir)
	}

	go w.run()
	return w, nil
}

func (w *Watcher) reload() {
	merged, err := LoadAll(w.internalDir, w.globalDir, w.projectDir)
	if err != nil {
		return
	}
	w.registry.replace(merged)
}

func (w *Watcher) run() {
	var debounce *time.Timer
	for {
		select {
		case <-w.stopCh:
			return
		case _, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(100*time.Millisecond, w.reload)
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopCh)
		_ = w.fsw.Close()
	})
}
