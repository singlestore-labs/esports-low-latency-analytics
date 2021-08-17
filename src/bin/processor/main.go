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

	config := &src.ProcessorConfig{}

	flag.IntVar(&config.Verbose, "verbose", 0, "Verbose level")
	flag.IntVar(&config.NumWorkers, "num-workers", runtime.NumCPU(), "Number of workers")
	flag.StringVar(&config.ReplayDir, "replay-dir", "", "Replay directory")

	flag.Parse()

	if len(config.ReplayDir) == 0 {
		log.Fatal("Replay directory and icon directory must be set")
	}

	log.SetFlags(log.Ldate | log.Ltime)

	dbConfig := src.SinglestoreConfigFromEnv()

	var db *src.Singlestore
	var err error
	for {
		db, err = src.NewSinglestore(dbConfig)
		if err != nil {
			log.Printf("unable to connect to SingleStore: %s; retrying...", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer db.Close()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting processor with %d workers", config.NumWorkers)

	workQueue := make(chan string)
	closeChannels := make([]chan struct{}, 0)
	wg := sync.WaitGroup{}

	for i := 0; i < config.NumWorkers; i++ {
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
