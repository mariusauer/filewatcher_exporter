package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	dirs         = kingpin.Flag("dirs", "Colon-separated list of directories to watch").Required().String()
	recursive    = kingpin.Flag("recursive", "Watch directories recursively").Bool()
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9150")

	lastWriteTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "file_last_write_timestamp_seconds",
			Help: "Last modification timestamp of any file in directory (event-driven)",
		},
		[]string{"directory"},
	)
)

func recordLastWrite(dir string) {
	lastWriteTimestamp.WithLabelValues(dir).Set(float64(time.Now().Unix()))
}

func addWatchers(w *fsnotify.Watcher, root string, recursive bool) error {
	if !recursive {
		if err := w.Add(root); err != nil {
			log.Printf("Failed to watch %s: %v", root, err)
			return err
		}
		log.Printf("Watching %s", root)
		return nil
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := w.Add(path); err != nil {
				log.Printf("Failed to watch %s: %v", path, err)
			} else {
				log.Printf("Watching %s", path)
			}
		}
		return nil
	})
}

func watchDirectory(root string, recursive bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	if err := addWatchers(watcher, root, recursive); err != nil {
		log.Fatalf("Failed to add watchers: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Update metric if file is written/created
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				recordLastWrite(root)
			}
			// If a new directory is created, add a watcher for it
			if recursive && event.Op&fsnotify.Create != 0 {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					if err := addWatchers(watcher, event.Name, recursive); err != nil {
						log.Printf("Failed to add watcher for new dir %s: %v", event.Name, err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error on %s: %v", root, err)
		}
	}
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	directories := filepath.SplitList(*dirs)

	registry := prometheus.NewRegistry()
	registry.MustRegister(lastWriteTimestamp)

	for _, d := range directories {
		go watchDirectory(d, *recursive)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	srv := &http.Server{
		Handler: mux,
	}

	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		log.Fatal(err)
	}
}
