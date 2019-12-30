package fsw

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var dev = false

// OnSave notify for a single file or directory, non-recursive.
// Monitor create and write operations.
// It runs callbacks in seperate goroutines.
// Files start with '.' are ignored.
func OnSave(ctx context.Context, dir string, fn func(basename string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	addAll := func(dir string) {
		// single file watch
		if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
			watcher.Add(dir)
			return
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if dev {
					log.Printf("[fsw] ADD %s", path)
				}
				err = watcher.Add(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	// watcher life time function
	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				name := filepath.Base(event.Name)
				// if strings.HasPrefix(name, ".") {
				// 	break
				// }

				if event.Op&fsnotify.Write == fsnotify.Write {
					fi, err := os.Stat(event.Name)
					if err != nil {
						break
					}
					if fi.IsDir() {
						if dev {
							log.Printf("[fsw] write dir %s (skip)", event.Name)
						}
						break
					}
					if dev {
						log.Printf("[fsw] write %s", event.Name)
					}
					if fn != nil {
						go fn(name)
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					fi, err := os.Stat(event.Name)
					if err != nil {
						break
					}
					if fi.IsDir() {
						addAll(event.Name)
						// break
					}
					if dev {
						log.Printf("[fsw] create %s", event.Name)
					}
					if fn != nil {
						go fn(name)
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					if dev {
						log.Printf("[fsw] remove %s", event.Name)
					}
					if fn != nil {
						go fn(name)
					}
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					if dev {
						log.Printf("[fsw] rename %s", event.Name)
					}
					if fn != nil {
						go fn(name)
					}
				} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					if dev {
						log.Printf("[fsw] chmod %s (skip)", event.Name)
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	addAll(dir)

	return nil
}
