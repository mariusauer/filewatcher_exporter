package main

import (
    stdlog "log"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/alecthomas/kingpin/v2"
    "github.com/fsnotify/fsnotify"
    kitlog "github.com/go-kit/log"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/exporter-toolkit/web"
    "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
    dirs = kingpin.Flag("dirs", "Colon-separated list of directories to watch").Default("/tmp").String()
    toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9000")

    lastWriteTimestamp = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "file_last_write_timestamp_seconds",
            Help: "Last modification timestamp of any file in directory (event-driven)",
        },
        []string{"directory"},
	)

	lastWriteAge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "file_last_write_age_seconds",
			Help: "Age in seconds since last modification in directory (event-driven)",
		},
		[]string{"directory"},
	)
)

// Shared state
type dirState struct {
	sync.Mutex
	lastWrite map[string]int64 // directory -> unix timestamp
}

func newDirState() *dirState {
	return &dirState{
		lastWrite: make(map[string]int64),
	}
}

func (ds *dirState) update(dir string) {
	ds.Lock()
	defer ds.Unlock()
	ts := time.Now().Unix()
	ds.lastWrite[dir] = ts
	lastWriteTimestamp.WithLabelValues(dir).Set(float64(ts))
}

func (ds *dirState) collect() {
	now := time.Now().Unix()
	ds.Lock()
	defer ds.Unlock()
	for d, ts := range ds.lastWrite {
		lastWriteAge.WithLabelValues(d).Set(float64(now - ts))
	}
}

// Add watchers recursively
func addWatchers(w *fsnotify.Watcher, root string) error {
    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
            if err := w.Add(path); err != nil {
                stdlog.Printf("Failed to watch %s: %v", path, err)
            } else {
                stdlog.Printf("Watching %s", path)
            }
        }
        return nil
    })
}

func watchRecursive(ds *dirState, root string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        stdlog.Fatalf("Failed to create watcher: %v", err)
    }
    defer watcher.Close()

    if err := addWatchers(watcher, root); err != nil {
        stdlog.Fatalf("Failed to add watchers: %v", err)
    }

    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            // Update metric if file is written/created
            if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
                ds.update(root)
            }
            // If a new directory is created, add a watcher for it
            if event.Op&fsnotify.Create != 0 {
                fi, err := os.Stat(event.Name)
                if err == nil && fi.IsDir() {
                    if err := addWatchers(watcher, event.Name); err != nil {
                        stdlog.Printf("Failed to add watcher for new dir %s: %v", event.Name, err)
                    }
                }
            }
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            stdlog.Printf("Watcher error on %s: %v", root, err)
        }
    }
}

func main() {
    kingpin.HelpFlag.Short('h')
    kingpin.Parse()
    directories := filepath.SplitList(*dirs)

    registry := prometheus.NewRegistry()
    registry.MustRegister(lastWriteTimestamp, lastWriteAge)

    ds := newDirState()

	// Launch a watcher per top-level directory
	for _, d := range directories {
		go watchRecursive(ds, d)
	}

	// Periodically refresh age metrics
	go func() {
		for {
			ds.collect()
			time.Sleep(10 * time.Second)
		}
	}()

    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

    srv := &http.Server{
        Handler: mux,
    }

    logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
    if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
        stdlog.Fatal(err)
    }
}
