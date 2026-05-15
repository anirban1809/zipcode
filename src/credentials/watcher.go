package credentials

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	filepath string
	store    *Store
	fsw      *fsnotify.Watcher
	stopOnce sync.Once
	stopCh   chan struct{}
}

func StartWatcher(store *Store, filePath string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	w := &Watcher{
		filepath: filePath,
		store:    store,
		fsw:      fsw,
		stopCh:   make(chan struct{}),
	}

	fsw.Add(filePath)
	go w.run()
	return w, nil
}

func (w *Watcher) reload() {
	w.store.Load()
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
