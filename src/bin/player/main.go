package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"src"

	"github.com/gin-gonic/gin"
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

	config := &src.PlayerConfig{}
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

	if config.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	server := src.NewReplayServer(config, db)
	server.RegisterRoutes(router)

	router.Run(fmt.Sprintf(":%d", config.Port))
}
