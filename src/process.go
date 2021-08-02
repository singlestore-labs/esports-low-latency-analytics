package src

import (
	"crypto/sha256"
	"encoding/binary"
	"html"
	"log"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

const (
	VerboseInfo  = 1
	VerboseDebug = 2
	VerboseSpam  = 3
)

// LoopTime returns a duration representing the time since the game started for a given loop index
func LoopTime(loop int) time.Duration {
	// 1 second = 16 loops => 1 loop = 1/16 second = 62,500,000 ns
	return time.Duration(loop * 62500000)
}

// UnitTag returns a unique value for each unit in the game
// SC2 reuses tagIndex for multiple units, the recycle value provides uniqueness
// note: these values are not comparable between two games - the UnitTag only
// serves as a unique id for the unit within a single game
func UnitTag(evt s2prot.Event) int64 {
	tagIndex := evt.Int("unitTagIndex")
	tagRecycle := evt.Int("unitTagRecycle")
	return (tagIndex << 18) + tagRecycle
}

type UnitInfo struct {
	PlayerId int
	UnitType string
	Alive    bool
}

func gameIDFromFileName(fileName string) int64 {
	data := sha256.Sum256([]byte(fileName))
	return int64(binary.BigEndian.Uint64(data[:8]))
}

func Run(env *ProcessorEnv, filename string) error {
	replay, err := rep.NewFromFile(filename)
	if err != nil {
		return err
	}

	if len(replay.Details.Players()) != 2 || len(replay.Metadata.Players()) != 2 {
		log.Printf("SKIP: found more than 2 players in replay: %s", filename)
		return nil
	}
	if replay.Header.Loops() >= 57600 {
		log.Printf("SKIP: replay longer than 1 hour (57600 loops): %s", filename)
		return nil
	}

	cleanFilename := strings.TrimPrefix(filename, env.ReplayDir+"/")
	gameID := gameIDFromFileName(cleanFilename)

	_, err = sq.
		Replace("games").
		SetMap(map[string]interface{}{
			"gameID":      gameID,
			"filename":    cleanFilename,
			"ts":          replay.Details.TimeUTC(),
			"loops":       replay.Header.Loops(),
			"durationSec": replay.Metadata.DurationSec(),
			"mapName":     replay.Metadata.Title(),
			"gameVersion": replay.Metadata.GameVersion(),
			"matchup":     replay.Details.Matchup(),
		}).
		RunWith(env.DB).
		Exec()
	if err != nil {
		return err
	}

	query := sq.Replace("players").RunWith(env.DB).Columns(
		"gameID", "playerID", "regionID", "realmID", "toonID", "name", "race", "opponentRace", "mmr", "apm", "result",
	)

	for i, player := range replay.Details.Players() {
		playerID := i + 1

		// we checked earlier that there were exactly 2 players and that both
		// player arrays are the same length
		metaPlayer := replay.Metadata.Players()[i]
		opponent := replay.Details.Players()[(i+1)%2]

		query = query.Values(
			gameID,
			playerID,
			player.Toon.RegionID(),
			player.Toon.RealmID(),
			player.Toon.ID(),
			html.UnescapeString(player.Name),
			player.Race().Name,
			opponent.Race().Name,
			metaPlayer.MMR(),
			metaPlayer.APM(),
			player.Result().Name,
		)
	}

	_, err = query.Exec()
	if err != nil {
		return err
	}

	statsLoader := NewLoader(env.DB, "playerstats", env.PlayerStatsSchema)
	defer func() {
		err := statsLoader.Close()
		if err != nil {
			log.Fatalf("PlayerStats Loader failed: %s", err)
		}
	}()

	buildCompLoader := NewLoader(env.DB, "buildcomp", env.BuildCompSchema)
	defer func() {
		err := buildCompLoader.Close()
		if err != nil {
			log.Fatalf("BuildComp Loader failed: %s", err)
		}
	}()

	writeBuildCompChange := func(loop int64, playerID int, unitType string, num int) error {
		// playerID 0 is used for map objects like mineral fields and so on
		if playerID == 0 {
			return nil
		}

		// ignore certain unitTypes which are cosmetic or not interesting
		if IgnoreUnitTypeRe.MatchString(unitType) {
			return nil
		}

		return buildCompLoader.Encode(&BuildCompChange{
			GameID:   gameID,
			PlayerID: playerID,
			LoopID:   loop,
			Kind:     unitType,
			Num:      num,
		})
	}

	unitMap := make(map[int64]*UnitInfo)

	for _, evt := range replay.TrackerEvts.Evts {
		switch evt.ID {
		case TrackerEvtIDPlayerStats:
			stats := evt.Structv("stats")
			statsLoader.Encode(&PlayerStats{
				GameID:   gameID,
				PlayerID: int(evt.Int("playerId")),
				LoopID:   evt.Loop(),
				Stats:    stats.String(),
			})
		case TrackerEvtIDUnitBorn:
			unitInfo := &UnitInfo{
				PlayerId: int(evt.Int("controlPlayerId")),
				UnitType: evt.Stringv("unitTypeName"),
				Alive:    true,
			}
			unitMap[UnitTag(evt)] = unitInfo

			if env.Verbose >= VerboseSpam {
				log.Printf("player %d created %s (%d)", unitInfo.PlayerId, unitInfo.UnitType, UnitTag(evt))
			}

			err := writeBuildCompChange(evt.Loop(), unitInfo.PlayerId, unitInfo.UnitType, 1)
			if err != nil {
				return err
			}
		case TrackerEvtIDUnitDied:
			tag := UnitTag(evt)
			unitInfo := unitMap[tag]

			if unitInfo.Alive {
				if env.Verbose >= VerboseSpam {
					log.Printf("player %d lost %s (%d)", unitInfo.PlayerId, unitInfo.UnitType, tag)
				}

				unitInfo.Alive = false
				err := writeBuildCompChange(evt.Loop(), unitInfo.PlayerId, unitInfo.UnitType, -1)
				if err != nil {
					return err
				}
			}
		case TrackerEvtIDUnitOwnerChange:
			loop := evt.Loop()
			tag := UnitTag(evt)
			unitInfo := unitMap[tag]
			newPlayerId := int(evt.Int("controlPlayerId"))

			// I have found cases where the unit changes ownership at the same instant as it dies.
			if unitInfo.Alive {
				if env.Verbose >= VerboseSpam {
					log.Printf("%s (%d) changed ownership from player %d to player %d", unitInfo.UnitType, tag, unitInfo.PlayerId, newPlayerId)
				}

				err := writeBuildCompChange(loop, unitInfo.PlayerId, unitInfo.UnitType, -1)
				if err != nil {
					return err
				}

				// change to new owner
				unitInfo.PlayerId = newPlayerId

				err = writeBuildCompChange(loop, unitInfo.PlayerId, unitInfo.UnitType, 1)
				if err != nil {
					return err
				}
			}
		case TrackerEvtIDUnitTypeChange:
			loop := evt.Loop()
			tag := UnitTag(evt)
			unitInfo := unitMap[tag]
			newType := evt.Stringv("unitTypeName")

			if unitInfo.Alive {
				if env.Verbose >= VerboseSpam {
					log.Printf("player %d's %s (%d) changed type to %s", unitInfo.PlayerId, unitInfo.UnitType, tag, newType)
				}

				err := writeBuildCompChange(loop, unitInfo.PlayerId, unitInfo.UnitType, -1)
				if err != nil {
					return err
				}

				// change to new type
				unitInfo.UnitType = newType

				err = writeBuildCompChange(loop, unitInfo.PlayerId, unitInfo.UnitType, 1)
				if err != nil {
					return err
				}
			}
		case TrackerEvtIDUpgrade:
			if env.Verbose >= VerboseSpam {
				log.Printf("player %d received upgrade %s", evt.Int("playerId"), evt.Stringv("upgradeTypeName"))
			}

			err = writeBuildCompChange(evt.Loop(), int(evt.Int("playerId")), evt.Stringv("upgradeTypeName"), 1)
			if err != nil {
				return err
			}
		case TrackerEvtIDUnitInit:
			unitInfo := &UnitInfo{
				PlayerId: int(evt.Int("controlPlayerId")),
				UnitType: evt.Stringv("unitTypeName"),
			}
			unitMap[UnitTag(evt)] = unitInfo

			if env.Verbose >= VerboseSpam {
				log.Printf("player %d started building %s (%d)", unitInfo.PlayerId, unitInfo.UnitType, UnitTag(evt))
			}

		case TrackerEvtIDUnitDone:
			unitInfo := unitMap[UnitTag(evt)]
			unitInfo.Alive = true

			if env.Verbose >= VerboseSpam {
				log.Printf("player %d finished building %s (%d)", unitInfo.PlayerId, unitInfo.UnitType, UnitTag(evt))
			}

			err := writeBuildCompChange(evt.Loop(), unitInfo.PlayerId, unitInfo.UnitType, 1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
