package processor

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Verbose    int `yaml:"verbose"`
	NumWorkers int `yaml:"numWorkers"`
}

func ParseConfigs(filenames []string) (*Config, error) {
	cfg := Config{}

	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&cfg)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func (c *Config) NewEnv(worker int) (*Env, error) {
	return &Env{WorkerID: worker}, nil
}
