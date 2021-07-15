package processor

import (
	"log"

	"github.com/icza/s2prot/rep"
)

func Run(env *Env, filename string) error {
	replay, err := rep.NewFromFile(filename)
	if err != nil {
		return err
	}

	for _, evt := range replay.TrackerEvts.Evts {
		log.Printf("found event: %s\n", evt.EvtType.Name)
	}

	return nil
}
