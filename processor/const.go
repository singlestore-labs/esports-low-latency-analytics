package processor

import "regexp"

const (
	/*
		PlayerStats attributes
			playerId
			stats
				scoreValueMineralsCurrent
				scoreValueVespeneCurrent
				scoreValueMineralsCollectionRate
				scoreValueVespeneCollectionRate
				scoreValueWorkersActiveCount
				scoreValueMineralsUsedInProgressArmy
				scoreValueMineralsUsedInProgressEconomy
				scoreValueMineralsUsedInProgressTechnology
				scoreValueVespeneUsedInProgressArmy
				scoreValueVespeneUsedInProgressEconomy
				scoreValueVespeneUsedInProgressTechnology
				scoreValueMineralsUsedCurrentArmy
				scoreValueMineralsUsedCurrentEconomy
				scoreValueMineralsUsedCurrentTechnology
				scoreValueVespeneUsedCurrentArmy
				scoreValueVespeneUsedCurrentEconomy
				scoreValueVespeneUsedCurrentTechnology
				scoreValueMineralsLostArmy
				scoreValueMineralsLostEconomy
				scoreValueMineralsLostTechnology
				scoreValueVespeneLostArmy
				scoreValueVespeneLostEconomy
				scoreValueVespeneLostTechnology
				scoreValueMineralsKilledArmy
				scoreValueMineralsKilledEconomy
				scoreValueMineralsKilledTechnology
				scoreValueVespeneKilledArmy
				scoreValueVespeneKilledEconomy
				scoreValueVespeneKilledTechnology
				scoreValueFoodUsed
				scoreValueFoodMade
				scoreValueMineralsUsedActiveForces
				scoreValueVespeneUsedActiveForces
				scoreValueMineralsFriendlyFireArmy
				scoreValueMineralsFriendlyFireEconomy
				scoreValueMineralsFriendlyFireTechnology
				scoreValueVespeneFriendlyFireArmy
				scoreValueVespeneFriendlyFireEconomy
				scoreValueVespeneFriendlyFireTechnology
	*/
	TrackerEvtIDPlayerStats = 0

	/*
		UnitBorn attributes
			unitTagIndex
			unitTagRecycle
			unitTypeName
			controlPlayerId
			upkeepPlayerId
			x
			y
			creatorUnitTagIndex
			creatorUnitTagRecycle
			creatorAbilityName
	*/
	TrackerEvtIDUnitBorn = 1

	/*
		UnitDied attributes
			unitTagIndex
			unitTagRecycle
			killerPlayerId
			x
			y
			killerUnitTagIndex
			killerUnitTagRecycle
	*/
	TrackerEvtIDUnitDied = 2

	/*
		UnitOwnerChange attributes
			unitTagIndex
			unitTagRecycle
			controlPlayerId
			upkeepPlayerId
	*/
	TrackerEvtIDUnitOwnerChange = 3

	/*
		UnitTypeChange attributes
			unitTagIndex
			unitTagRecycle
			unitTypeName
	*/
	TrackerEvtIDUnitTypeChange = 4

	/*
		Upgrade attributes
			playerId
			upgradeTypeName
			count
	*/
	TrackerEvtIDUpgrade = 5

	/*
		UnitInit attributes
			unitTagIndex
			unitTagRecycle
			unitTypeName
			controlPlayerId
			upkeepPlayerId
			x
			y
	*/
	TrackerEvtIDUnitInit = 6

	/*
		UnitDone attributes
			unitTagIndex
			unitTagRecycle
	*/
	TrackerEvtIDUnitDone = 7
)

var (
	// IgnoreUnitTypeRe is a regex which matches unit types which should be ignored
	IgnoreUnitTypeRe = regexp.MustCompile(`^(Beacon|RewardDance|Spray|LoadOutSpray|GameHeartActive)`)
)
