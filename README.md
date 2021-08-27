# Reference architecture using SingleStore for realtime esports analytics

The goal of this reference architecture is to build an esports dashboard displaying
low-latency analytics for commentators.

## Background

Extremely experienced commentators can reflect on famous events and games, but often finding complex
patterns when multiple events are happening at the same time is difficult.

In order to display high quality analytics, we need real data. This reference architecture uses
a data set can be found at the [UCI ML repository](http://archive.ics.uci.edu/ml/datasets/SkillCraft1+Master+Table+Dataset).

StarCraft 2 (SC2) is a realtime strategy (RTS) video game developed and published by Blizzard Entertainment
with esports matches beginning in 2010. SC2 has a large community and several professional leagues.
There are three races with unique units and strategies: terran, zerg, protoss. The goal of each match
is to destroy all of the opponents buildings or create a situation where there is no chance of winning and
the opponent forfeits.

The data set includes 10,000
StarCraft 2 replays where the players are top ranked professional players. If you follow SC2, you may recognize many of the names.

## Analytics

We have a “buildcomp“ table, short for build composition. For each game, for each player, in one game loop (1/16th of second). We are looking for the kind and number of units, e.g. kind = zergling, num = 1.

Buildcomp reflects the delta of the kinds of events that are occurring.

The compvecs table is a prebuilt result of all of the games. This produces a SingleStore floating point vector that reflects the composition at that point.

## Demo

On the home screen, each card represents a unique StartCraft 2 replay from professional players.

1.  Select Map: Catalyst Zest (Protoss) vs Impact (Zerg)
1.  Hit the play button in the top-right hand corner.
    The first timeline at the top represents a live game in realtime. We are monitoring for new events.
    The timelines below this one are similar games that match the realtime match's build composition.

    To start, almost every Protoss player will create the same initial
    units. The first fork begins at about **1m40s**.

    There are two key strategies that professional Protoss players employ,
    Early Forge might lead to a Photon Cannon Rush or an Assimilator.

    This game is about managing resources. A building unit consumes resources of crystals and gas.
    How those resources are allocated is key to late game strategies.

    In this case, the player has chosen an assimilator. As soon as the assimilator has been chosen, this eliminated the early forge strategy. Notice how the timelines have changed.
1.  A commentator might ask something like "When will the player make their Gateway?"
    or "Will there be a second early assimilator?"
    No. We see that this is a very standard Gateway build.
    Next choices include building an assimilator before building Adepts (a warrior type).
1.  At **3m13s**, checkout that we found a previous game from the same player.
    It's very informative to see a given player's past strategies.
    This system is very effective at predicting what happens next.
1.  Look for the potential of a Stargate at **4m10s**.

As we can see, a commentator can look at this dashboard with specific context and provide
informative commentatory for things the audience can look for.

With additional features, we could even have a fully automated commentator.

## Technical details

All of this data is simulated.

Per event we compute <60ms
k-nearest neighbor search of 20
million build composition vectors.
Each one of those vectors, is a small window of one of the 10,000 games
we loaded into our database.

We search for the 5 most similar games.

This is done by dynamically producing the buildcomp vector of the realtime game,
then we look for that same vector in the 20 million vector dataset.

We aggregate down to the 5 most similar games.
The second timeline is the most similar game.

Right now, we don't treat units with different weights, so there are definitely pieces that could be improved, but the bones are there.

We are running potentially hundreds of simularity searches within milliseconds. Because of the power of singlestore, all of this is done in realtime.

## Processor

**Player stats timeseries**

Emit general stats per player every 10 seconds (TrackerEvent -> PlayerStats)

**Army composition time series**

```
spawning = {}

for each TrackerEvent:
    switch ID
        case UnitBorn: emit (playerIdx, unitType, +1)
        case UnitInit: spawning[unitTag] = unitType
        case UnitDone: emit (playerIdx, spawning[unitTag], +1)
        case UnitDied: emit (playerIdx, unitType, -1)
        case UnitOwnerChange:
            emit (fromPlayerIdx, fromUnitType, -1)
            emit (toPlayerIdx, fromUnitType, +1)
        case UnitTypeChange:
            emit (playerIdx, fromUnitType, -1)
            emit (playerIdx, toUnitType, +1)
```

## Sources

Thompson JJ, Blair MR, Chen L, Henrey AJ (2013) Video Game Telemetry as a Critical Tool in the Study of Complex Skill Learning. PLoS ONE 8(9): e75129. [Web Link]

http://archive.ics.uci.edu/ml/datasets/SkillCraft1+Master+Table+Dataset
