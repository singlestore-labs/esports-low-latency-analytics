package src

import (
	"fmt"
)

type Event struct {
	PlayerID int   `json:"playerid"`
	LoopID   int64 `json:"loopid"`

	Kind string `json:"kind"`
	Num  int    `json:"num"`
}

type Stats struct {
	PlayerID int   `json:"playerid"`
	LoopID   int64 `json:"loopid"`

	FoodMade               int `json:"foodMade"`
	FoodUsed               int `json:"foodUsed"`
	MineralsCollectionRate int `json:"mineralsCollectionRate"`
	MineralsCurrent        int `json:"mineralsCurrent"`
	VespeneCollectionRate  int `json:"vespeneCollectionRate"`
	VespeneCurrent         int `json:"vespeneCurrent"`
}

type Timeline struct {
	GameID int64   `json:"gameid"`
	Events []Event `json:"events"`
	Stats  []Stats `json:"stats"`
}

func LoadTimeline(db *Singlestore, gameID int64) (*Timeline, error) {
	events := make([]Event, 0)
	stats := make([]Stats, 0)

	err := db.Select(&events, `
		select playerid, loopid, kind, num
		from buildcomp
		where gameid = ?
		order by loopid asc, kind
	`, gameID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for gameid: %d", gameID)
	}

	err = db.Select(&stats, `
		select
			playerid, loopid,
			foodmade, foodused,
			mineralscollectionrate, mineralscurrent,
			vespenecollectionrate, vespenecurrent
		from playerstats
		where gameid = ?
		order by loopid asc, playerid
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
		Stats:  stats,
	}, nil
}

func (t *Timeline) MaxLoopID() int64 {
	return t.Events[len(t.Events)-1].LoopID
}
