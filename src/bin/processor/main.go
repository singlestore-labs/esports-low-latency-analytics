package main

import (
	"errors"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"src"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	configPaths := src.FlagStringSlice{}
	flag.Var(&configPaths, "config", "path to the config file; can be provided multiple times, files will be merged in the order provided")
	flag.Parse()

	if len(configPaths) == 0 {
		configPaths.Set("config.toml")
	}

	log.SetFlags(log.Ldate | log.Ltime)

	config := &src.ProcessorConfig{}
	err := src.LoadTOMLFiles(config, []string(configPaths))
	if err != nil {
		log.Fatalf("unable to load config files: %v; error: %+v", configPaths, err)
	}

	var db *src.Singlestore
	for {
		db, err = src.NewSinglestore(config.Singlestore)
		if err != nil {
			log.Printf("unable to connect to SingleStore: %s; retrying...", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer db.Close()

	numWorkers := runtime.NumCPU()
	if config.NumWorkers != 0 {
		numWorkers = config.NumWorkers
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting processor with %d workers", numWorkers)

	workQueue := make(chan string)
	closeChannels := make([]chan struct{}, 0)
	wg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		closeCh := make(chan struct{})
		closeChannels = append(closeChannels, closeCh)

		env := src.NewProcessorEnv(i, config, db)

		go func() {
			defer wg.Done()

			for {
				select {
				case file := <-workQueue:
					log.Printf("processing file %s", file)
					time.Sleep(time.Second)
					err := src.Run(env, file)
					if err != nil {
						log.Fatalf("error processing file %s: %v", file, err)
					}
				case <-closeCh:
					return
				}
			}
		}()
	}

	errShutdown := errors.New("shutdown")

	err = filepath.Walk(config.ReplayDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		basename := filepath.Base(path)
		if !strings.HasSuffix(basename, ".SC2Replay") || strings.HasPrefix(basename, ".") {
			return nil
		}

		select {
		case workQueue <- path:
		case sig := <-signals:
			log.Printf("received shutdown signal: %s", sig)
			return errShutdown
		}
		return nil
	})

	if err != nil && err != errShutdown {
		log.Fatalf("error while processing %s: %v", config.ReplayDir, err)
	}

	for _, ch := range closeChannels {
		close(ch)
	}

	wg.Wait()

	if err == nil {
		now := time.Now()
		log.Printf("starting postprocess() at %s", now)
		if _, err := db.Exec("CALL postprocess()"); err != nil {
			log.Fatalf("postprocess() failed: %s", err)
		}
		log.Printf("postprocess() finished in %s", time.Since(now))
	}
}
