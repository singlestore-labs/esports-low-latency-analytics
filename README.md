# Reference Architecture using SingleStore for real-time Esports analytics

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
