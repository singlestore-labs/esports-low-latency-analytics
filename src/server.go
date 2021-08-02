package src

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReplayServer struct {
	Config *PlayerConfig
	DB     *Singlestore

	replays map[int64]*ReplaySimulator
}

func NewReplayServer(config *PlayerConfig, db *Singlestore) *ReplayServer {
	return &ReplayServer{
		Config:  config,
		DB:      db,
		replays: make(map[int64]*ReplaySimulator),
	}
}

func (s *ReplayServer) RegisterRoutes(router gin.IRouter) error {
	router.GET("/api/replays", s.ListReplays)
	router.GET("/api/replays/start", s.StartReplay)
	router.GET("/api/replays/active", s.ActiveReplays)
	router.GET("/api/replays/stop", s.StopReplay)
	return nil
}

func (s *ReplayServer) ListReplays(c *gin.Context) {
	params := struct {
		Matchup string `form:"matchup"`
		Player  string `form:"player"`
		Limit   int    `form:"limit"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Limit == 0 {
		params.Limit = 100
	}

	out := []struct {
		GameID   int64
		Filename string
		Mapname  string
		P1Name   string
		P1Race   string
		P1Result string
		P2Name   string
		P2Race   string
		P2Result string
	}{}

	query, args, err := s.DB.BindNamed(`
		select
			games.gameid, games.filename, games.mapname,
			p1.name p1name, p1.race p1race, p1.result p1result,
			p2.name p2name, p2.race p2race, p2.result p2result
		from games, players p1, players p2
		where
			games.gameid = p1.gameid and p1.playerid = 1
			and games.gameid = p2.gameid and p2.playerid = 2
			and (:matchup = "" or games.matchup = :matchup)
			and (
				:player = ""
				or p1.name like concat("%",:player,"%")
				or p2.name like concat("%",:player,"%")
			)
		limit :limit
	`, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = s.DB.Select(&out, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, out)
}

func (s *ReplayServer) StartReplay(c *gin.Context) {
	params := struct {
		GameID    int64 `form:"gameid"`
		StartLoop int64 `form:"startloop"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exists := s.replays[params.GameID]; exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game already started"})
		return
	}

	replay, err := StartReplay(s.DB, params.GameID, params.StartLoop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.replays[replay.GameID] = replay

	c.JSON(200, gin.H{"started": true})
}

func (s *ReplayServer) ActiveReplays(c *gin.Context) {
	out := make([]int64, 0, len(s.replays))
	for k := range s.replays {
		out = append(out, k)
	}
	c.JSON(200, out)
}

func (s *ReplayServer) StopReplay(c *gin.Context) {
	params := struct {
		GameID int64 `form:"gameid"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exists := s.replays[params.GameID]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not running"})
		return
	}

	s.replays[params.GameID].Stop()
	delete(s.replays, params.GameID)

	c.JSON(200, gin.H{"stopped": true})
}
