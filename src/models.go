package src

type PlayerStats struct {
	GameID   int64
	PlayerID int
	LoopID   int64
	// Stats is a JSON string
	Stats string
}

type BuildCompChange struct {
	GameID   int64
	PlayerID int
	LoopID   int64

	Kind string
	Num  int
}
