package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"src"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config := &src.PlayerConfig{}

	flag.IntVar(&config.Verbose, "verbose", 0, "Verbose level")
	flag.StringVar(&config.ReplayDir, "replay-dir", "", "Replay directory")
	flag.StringVar(&config.IconDir, "icon-dir", "", "Icon directory")
	flag.IntVar(&config.Port, "port", 8000, "Port")

	flag.Parse()

	if len(config.ReplayDir) == 0 || len(config.IconDir) == 0 {
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

	if config.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		MaxAge:          12 * time.Hour,
	}))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	server := src.NewReplayServer(config, db)
	server.RegisterRoutes(router)

	router.Run(fmt.Sprintf(":%d", config.Port))
}
