package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"processor"
)

type FlagStringSlice []string

func (f *FlagStringSlice) String() string {
	return "[]string"
}

func (f *FlagStringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	configPaths := FlagStringSlice{}
	cpuprofile := ""

	flag.Var(&configPaths, "config", "path to the config file; can be provided multiple times, files will be merged in the order provided")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to specified path")

	flag.Parse()

	if len(configPaths) == 0 {
		configPaths.Set("config.yaml")
	}

	files := flag.Args()
	if len(files) < 1 {
		fmt.Printf("usage: %s [flags] REPLAY_FILE REPLAY_FILE...", os.Args[0])
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime)

	config, err := processor.ParseConfigs([]string(configPaths))
	if err != nil {
		log.Fatalf("unable to load config files: %v; error: %+v", configPaths, err)
	}

	if cpuprofile != "" {
		// disable logging and lower verbosity during profile
		log.SetOutput(ioutil.Discard)
		config.Verbose = 0
	}

	log.Printf("processing %d files", len(files))

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// start the cpu profile after we initialize everything so we measure the
	// main simulation routines
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	numWorkers := runtime.NumCPU()
	if config.NumWorkers != 0 {
		numWorkers = config.NumWorkers
	}

	workQueue := make(chan string)
	closeChannels := make([]chan struct{}, 0)
	wg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		closeCh := make(chan struct{})
		closeChannels = append(closeChannels, closeCh)

		env, err := config.NewEnv(i)
		if err != nil {
			log.Fatalf("unable to create environment: %v", err)
		}

		go func() {
			defer wg.Done()

			for {
				select {
				case file := <-workQueue:
					log.Printf("processing file %s", file)
					time.Sleep(time.Second)
					err := processor.Run(env, file)
					if err != nil {
						log.Fatalf("error processing file %s: %v", file, err)
					}
				case <-closeCh:
					return
				}
			}
		}()
	}

outer:
	for _, file := range files {
		select {
		case workQueue <- file:
		case sig := <-signals:
			log.Printf("received shutdown signal: %s", sig)
			break outer
		}
	}

	for _, ch := range closeChannels {
		close(ch)
	}

	wg.Wait()
}
