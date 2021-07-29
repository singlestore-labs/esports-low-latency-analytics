package processor

import (
	"log"

	"github.com/hamba/avro"
)

type Env struct {
	WorkerID int
	DB       *Singlestore
	Verbose  int

	PlayerStatsSchema avro.Schema
	BuildCompSchema   avro.Schema
}

func NewEnv(workerID int, config *Config, db *Singlestore) *Env {
	statsSchema, err := AvroSchemaFromStruct(&PlayerStats{})
	if err != nil {
		log.Fatalf("failed to convert PlayerStats to avro schema: %s", err)
	}
	buildCompSchema, err := AvroSchemaFromStruct(&BuildCompChange{})
	if err != nil {
		log.Fatalf("failed to convert BuildCompChange to avro schema: %s", err)
	}

	return &Env{
		WorkerID: workerID,
		DB:       db,
		Verbose:  config.Verbose,

		PlayerStatsSchema: statsSchema,
		BuildCompSchema:   buildCompSchema,
	}
}
