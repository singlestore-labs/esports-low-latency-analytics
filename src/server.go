package src

import (
	"fmt"
	"net/http"
	"path"

	"cuelang.org/go/pkg/strconv"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ReplayServer struct {
	Config *PlayerConfig
	DB     *Singlestore
}

func NewReplayServer(config *PlayerConfig, db *Singlestore) *ReplayServer {
	return &ReplayServer{
		Config: config,
		DB:     db,
	}
}

func (s *ReplayServer) RegisterRoutes(router gin.IRouter) error {
	router.GET("/api/replays", s.ListReplays)
	router.GET("/api/replays/:gameid", s.GetReplay)
	router.GET("/api/replays/:gameid/timeline", s.GetReplayTimeline)
	router.GET("/api/replays/:gameid/similar", s.GetSimilarReplays)
	router.GET("/api/icon/:kind", s.GetIcon)
	return nil
}

type ReplayMeta struct {
	GameID   string `json:"gameid"`
	Filename string `json:"filename"`
	Mapname  string `json:"mapname"`
	Loops    int64  `json:"loops"`
	P1Name   string `json:"p1Name"`
	P1Race   string `json:"p1Race"`
	P1Result string `json:"p1Result"`
	P2Name   string `json:"p2Name"`
	P2Race   string `json:"p2Race"`
	P2Result string `json:"p2Result"`
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

func (s *ReplayServer) GetReplay(c *gin.Context) {
	gameid, err := ParamInt64(c, "gameid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out := &ReplayMeta{}
	err = s.DB.Get(out, `
		select
			games.gameid, games.filename, games.mapname, games.loops,
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
		order by if(games.gameid = -5280689129783593904, NULL, games.gameid) desc nulls first
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

func ParamInt64(c *gin.Context, name string) (int64, error) {
	s := c.Param(name)
	if s == "" {
		return 0, errors.Errorf("%s is required", name)
	}
	return strconv.ParseInt(s, 10, 64)
}

func (s *ReplayServer) GetReplayTimeline(c *gin.Context) {
	gameid, err := ParamInt64(c, "gameid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeline, err := LoadTimeline(s.DB, gameid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, timeline)
}

type SimilarGamePoint struct {
	GameID   string `json:"gameid"`
	PlayerID int    `json:"playerid"`
	LoopID   int64  `json:"loop"`
}

func (s *ReplayServer) GetSimilarReplays(c *gin.Context) {
	gameid, err := ParamInt64(c, "gameid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := struct {
		PlayerID int   `form:"playerid"`
		LoopID   int64 `form:"loop"`
		Lag      int64 `form:"lag"`
		Limit    int   `form:"limit"`
	}{}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playerInfo := struct {
		Race         string
		OpponentRace string
	}{}

	err = s.DB.Get(&playerInfo, `
		select race, opponentrace
		from players
		where gameID = ? and playerID = ?
	`, gameid, params.PlayerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := []SimilarGamePoint{}
	err = s.DB.Select(&out, `
			select distinct gameid, playerid, loopid
			from similarGamePoints(?, ?, ?, ?, ?, ?, ?)
		`,
		gameid, params.PlayerID, playerInfo.Race, playerInfo.OpponentRace,
		params.LoopID, params.Lag, params.Limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, out)
}
