DROP DATABASE IF EXISTS sc2;
CREATE DATABASE sc2;
USE sc2;

CREATE REFERENCE TABLE games (
    gameID BIGINT PRIMARY KEY NOT NULL,
    filename TEXT NOT NULL,
    ts DATETIME NOT NULL,

    loops BIGINT NOT NULL,
    durationSec DOUBLE NOT NULL,
    mapName TEXT NOT NULL,
    gameVersion TEXT NOT NULL,
    matchup TEXT NOT NULL
);

CREATE TABLE players (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,

    regionID BIGINT NOT NULL,
    realmID BIGINT NOT NULL,
    toonID BIGINT NOT NULL,

    name TEXT NOT NULL,
    race TEXT NOT NULL,
    opponentRace TEXT NOT NULL,

    mmr DOUBLE NOT NULL,
    apm DOUBLE NOT NULL,
    result TEXT NOT NULL,

    PRIMARY KEY (gameID, playerID),
    SORT KEY (gameID, playerID)
);

CREATE TABLE playerstats (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,
    loopID BIGINT NOT NULL,
    stats JSON NOT NULL,

    SORT KEY (gameID, loopID),
    SHARD (gameID, playerID)
);

CREATE TABLE buildcomp (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,
    loopID BIGINT NOT NULL,

    kind TEXT NOT NULL COLLATE "utf8_bin",
    num INT NOT NULL,

    SORT KEY (gameID, playerID, loopID),
    SHARD (gameID, playerID)
);

CREATE TABLE compvecs (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,
    race TEXT NOT NULL,
    opponentRace TEXT NOT NULL,
    loopID BIGINT NOT NULL,
    vec LONGBLOB NOT NULL,

    SORT KEY (gameID, playerID),
    SHARD (gameID, playerID),

    KEY (race, opponentRace)
);

CREATE ROWSTORE TABLE livebuildcomp (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,
    loopID BIGINT NOT NULL,

    kind TEXT NOT NULL COLLATE "utf8_bin",
    num INT NOT NULL,

    KEY (gameID, playerID, loopID),
    SHARD (gameID, playerID)
);

/*
-- quickly create a new version of a table
create table temp (
);
insert into temp select * from compvecs;
drop table compvecs;
alter table temp rename to compvecs;
*/