package src

import (
	"log"

	"github.com/hamba/avro"
)

type ProcessorEnv struct {
	WorkerID  int
	DB        *Singlestore
	Verbose   int
	ReplayDir string

	PlayerStatsSchema avro.Schema
	BuildCompSchema   avro.Schema
}

func NewProcessorEnv(workerID int, config *ProcessorConfig, db *Singlestore) *ProcessorEnv {
	statsSchema, err := AvroSchemaFromStruct(&PlayerStats{})
	if err != nil {
		log.Fatalf("failed to convert PlayerStats to avro schema: %s", err)
	}
	buildCompSchema, err := AvroSchemaFromStruct(&BuildCompChange{})
	if err != nil {
		log.Fatalf("failed to convert BuildCompChange to avro schema: %s", err)
	}

	return &ProcessorEnv{
		WorkerID:  workerID,
		DB:        db,
		Verbose:   config.Verbose,
		ReplayDir: config.ReplayDir,

		PlayerStatsSchema: statsSchema,
		BuildCompSchema:   buildCompSchema,
	}
}
