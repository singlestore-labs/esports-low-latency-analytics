package src

type PlayerStats struct {
	GameID   int64
	PlayerID int
	LoopID   int64

	FoodMade               int
	FoodUsed               int
	MineralsCollectionRate int
	MineralsCurrent        int
	VespeneCollectionRate  int
	VespeneCurrent         int
}

type BuildCompChange struct {
	GameID   int64
	PlayerID int
	LoopID   int64

	Kind string
	Num  int
}
