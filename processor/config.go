package processor

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Verbose    int
	NumWorkers int

	Singlestore SinglestoreConfig
}

func ParseConfigs(filenames []string) (*Config, error) {
	cfg := Config{}

	for _, filename := range filenames {
		_, err := toml.DecodeFile(filename, &cfg)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}