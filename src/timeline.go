package src

import (
	"fmt"
)

type Event struct {
	PlayerID int
	LoopID   int64
	Kind     string
	Num      int
}

type Timeline struct {
	GameID int64
	Events []Event
}

func LoadTimeline(db *Singlestore, gameID int64) (*Timeline, error) {
	events := make([]Event, 0)

	err := db.Select(&events, `
		select playerid, loopid, kind, num
		from buildcomp
		where gameid = ?
		order by loopid asc
	`, gameID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for gameid: %d", gameID)
	}

	return &Timeline{
		GameID: gameID,
		Events: events,
	}, nil
}

func (t *Timeline) MaxLoopID() int64 {
	return t.Events[len(t.Events)-1].LoopID
}
