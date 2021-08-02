package src

import (
	"fmt"
	"net/http"
	"path"

	"cuelang.org/go/pkg/strconv"
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
	router.GET("/api/replays/:gameid", s.GetReplay)
	router.GET("/api/replays/:gameid/status", s.GetReplayStatus)
	router.GET("/api/replays/:gameid/start", s.StartReplay)
	router.GET("/api/replays/:gameid/stop", s.StopReplay)
	router.GET("/api/replays/:gameid/timeline", s.GetReplayTimeline)
	router.GET("/api/icon/:kind", s.GetIcon)
	router.GET("/api/replays/active", s.ActiveReplays)
	return nil
}

type ReplayMeta struct {
	GameID   string `json:"gameid"`
	Filename string `json:"filename"`
	Mapname  string `json:"mapname"`
	P1Name   string `json:"p1Name"`
	P1Race   string `json:"p1Race"`
	P1Result string `json:"p1Result"`
	P2Name   string `json:"p2Name"`
	P2Race   string `json:"p2Race"`
	P2Result string `json:"p2Result"`
}

type ReplayStatus struct {
	Active bool  `json:"active"`
	Loop   int64 `json:"loop,omitempty"`
}

func (s *ReplayServer) GetIcon(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}
	out := struct {
		Icon string
	}{}
	err := s.DB.Get(&out, `
		select icon from kind2icon where kind = ?
	`, kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(path.Join(s.Config.IconDir, fmt.Sprintf("%s.png", out.Icon)))
}

func (s *ReplayServer) GetReplayStatus(c *gin.Context) {
	gameid := c.Param("gameid")
	if gameid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid is required"})
		return
	}
	gameidInt, err := strconv.ParseInt(gameid, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid must be a bigint"})
		return
	}

	for k, r := range s.replays {
		if k == gameidInt {
			c.JSON(http.StatusOK, ReplayStatus{
				Active: true,
				Loop:   r.CurrentLoop(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, ReplayStatus{
		Active: false,
		Loop:   0,
	})
}

func (s *ReplayServer) GetReplay(c *gin.Context) {
	gameid := c.Param("gameid")
	if gameid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid is required"})
		return
	}

	out := &ReplayMeta{}
	err := s.DB.Get(out, `
		select
			games.gameid, games.filename, games.mapname,
			p1.name p1name, p1.race p1race, p1.result p1result,
			p2.name p2name, p2.race p2race, p2.result p2result
		from games, players p1, players p2
		where games.gameid = ?
			and games.gameid = p1.gameid and p1.playerid = 1
			and games.gameid = p2.gameid and p2.playerid = 2
	`, gameid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, out)
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

	out := []ReplayMeta{}

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
		order by games.gameid desc
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
	gameid := c.Param("gameid")
	if gameid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid is required"})
		return
	}
	gameidInt, err := strconv.ParseInt(gameid, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid must be a bigint"})
		return
	}

	params := struct {
		StartLoop int64 `form:"startloop"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exists := s.replays[gameidInt]; exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game already started"})
		return
	}

	replay, err := StartReplay(s.DB, gameidInt, params.StartLoop)
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
	gameid := c.Param("gameid")
	if gameid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid is required"})
		return
	}
	gameidInt, err := strconv.ParseInt(gameid, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid must be a bigint"})
		return
	}

	if _, exists := s.replays[gameidInt]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not running"})
		return
	}

	s.replays[gameidInt].Stop()
	delete(s.replays, gameidInt)

	c.JSON(200, gin.H{"stopped": true})
}

func (s *ReplayServer) GetReplayTimeline(c *gin.Context) {
	gameid := c.Param("gameid")
	if gameid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid is required"})
		return
	}
	gameidInt, err := strconv.ParseInt(gameid, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameid must be a bigint"})
		return
	}

	params := struct {
		MinLoopID int64 `form:"minloop"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeline, err := LoadTimeline(s.DB, gameidInt, params.MinLoopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, timeline.Events)
}
