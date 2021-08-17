package src

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type ProcessorConfig struct {
	Verbose     int
	NumWorkers  int
	ReplayDir   string
	Singlestore SinglestoreConfig
}

type PlayerConfig struct {
	Verbose     int
	ReplayDir   string
	IconDir     string
	Port        int
	GinMode     string
	Singlestore SinglestoreConfig
}

type SinglestoreConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func LoadTOMLFiles(out interface{}, filenames []string) error {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Printf("toml file `%s` not found, skipping", filename)
			continue
		}
		_, err := toml.DecodeFile(filename, out)
		if err != nil {
			return err
		}
	}
	return nil
}
