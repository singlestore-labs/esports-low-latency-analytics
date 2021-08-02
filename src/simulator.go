package src

import (
	"log"
	"sync/atomic"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type ReplaySimulator struct {
	DB        *Singlestore
	GameID    int64
	Timeline  *Timeline
	StartLoop int64
	loop      int64
	stop      chan struct{}
}

func StartReplay(db *Singlestore, gameid int64, startloop int64) (*ReplaySimulator, error) {
	timeline, err := LoadTimeline(db, gameid, 0)
	if err != nil {
		return nil, err
	}

	sim := &ReplaySimulator{
		DB:        db,
		GameID:    gameid,
		Timeline:  timeline,
		StartLoop: startloop,
		loop:      0,
		stop:      make(chan struct{}),
	}

	go sim.Run()

	return sim, nil
}

func (r *ReplaySimulator) CurrentLoop() int64 {
	return atomic.LoadInt64(&r.loop)
}

func (r *ReplaySimulator) Run() {
	// clear the current game from livebuildcomp
	_, err := sq.Delete("livebuildcomp").Where(sq.Eq{"gameid": r.GameID}).RunWith(r.DB).Exec()
	if err != nil {
		log.Fatalf("Error deleting game %d from livebuildcomp: %s", r.GameID, err)
	}

	builder := sq.StatementBuilder.RunWith(r.DB)
	batch := builder.Insert("livebuildcomp")

	maxLoopID := r.Timeline.MaxLoopID()
	evtIdx := 0

	timer := time.NewTicker(LoopTime(1))

	for loop := int64(0); loop < maxLoopID; loop++ {
		atomic.StoreInt64(&r.loop, loop)

		// emit events
		emitBatch := false
		for {
			evt := r.Timeline.Events[evtIdx]
			if evt.LoopID > loop {
				break
			}

			batch = batch.Values(r.GameID, evt.PlayerID, evt.LoopID, evt.Kind, evt.Num)
			emitBatch = true
			evtIdx++
		}

		if emitBatch {
			_, err := batch.Exec()
			if err != nil {
				log.Fatalf("Error inserting batch into livebuildcomp: %s", err)
			}
			batch = builder.Insert("livebuildcomp")
		}

		if loop > r.StartLoop {
			select {
			case <-timer.C:
			case <-r.stop:
				return
			}
		}
	}
}

func (r *ReplaySimulator) Stop() {
	close(r.stop)
}
