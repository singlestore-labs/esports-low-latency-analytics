package src

import (
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
	GinMode     string
	Port        int
	Singlestore SinglestoreConfig
}

func LoadTOMLFiles(out interface{}, filenames []string) error {
	for _, filename := range filenames {
		_, err := toml.DecodeFile(filename, out)
		if err != nil {
			return err
		}
	}
	return nil
}
