package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"evidence-bridge/internal/log"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	Events    chan string
	mu        sync.Mutex
	timers    map[string]*time.Timer
}

// NewWatcher は、デバウンス処理を備えたファイル監視を開始します。
// Windowsでの複数イベント重なり（Create+Write+Chmodなど）を排除します。
func NewWatcher(dir string) (*Watcher, error) {
	// ディレクトリ作成
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("監視対象ディレクトリ作成失敗: %w", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = w.Add(dir)
	if err != nil {
		w.Close()
		return nil, err
	}

	res := &Watcher{
		fsWatcher: w,
		Events:    make(chan string, 100),
		timers:    make(map[string]*time.Timer),
	}

	go res.listen()
	return res, nil
}

func (w *Watcher) listen() {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			
			// 監視対象外の操作（削除、名前変更など）は無視
			if event.Op&fsnotify.Create != fsnotify.Create && event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			// デバウンス処理: 200ms以内に同一ファイルへの複数のイベントが来た場合、タイマーをリセット
			w.mu.Lock()
			if t, ok := w.timers[event.Name]; ok {
				t.Stop()
			}

			w.timers[event.Name] = time.AfterFunc(300*time.Millisecond, func() {
				info, err := os.Stat(event.Name)
				if err == nil && !info.IsDir() {
					w.Events <- event.Name
				}
				
				w.mu.Lock()
				delete(w.timers, event.Name)
				w.mu.Unlock()
			})
			w.mu.Unlock()

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			log.Warn("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) Close() error {
	w.mu.Lock()
	for _, t := range w.timers {
		t.Stop()
	}
	w.mu.Unlock()
	return w.fsWatcher.Close()
}

func ScanExistingFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
