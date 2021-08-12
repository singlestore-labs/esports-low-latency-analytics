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

CREATE ROWSTORE REFERENCE TABLE uniquekind (
    kind TEXT NOT NULL COLLATE "utf8_bin",
    PRIMARY KEY (kind)
);

CREATE ROWSTORE REFERENCE TABLE kind2icon (
    kind TEXT NOT NULL COLLATE "utf8_bin",
    icon TEXT NOT NULL,
    PRIMARY KEY (kind)
);

LOAD DATA INFILE '/data/kind2icon.csv'
SKIP DUPLICATE KEY ERRORS
INTO TABLE kind2icon
FIELDS TERMINATED BY ','
LINES TERMINATED BY '\n';

CREATE TABLE players (
    gameID BIGINT NOT NULL,
    playerID INT NOT NULL,

    regionID BIGINT NOT NULL,
    realmID BIGINT NOT NULL,
    toonID BIGINT NOT NULL,

    name TEXT NOT NULL,
    race TEXT NOT NULL COLLATE "utf8_bin",
    opponentRace TEXT NOT NULL COLLATE "utf8_bin",

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
    race TEXT NOT NULL COLLATE "utf8_bin",
    opponentRace TEXT NOT NULL COLLATE "utf8_bin",

    loopID BIGINT NOT NULL,
    loopLag BIGINT,

    vec LONGBLOB NOT NULL,

    SORT KEY (race, opponentRace) with (columnstore_segment_rows=200000),
    SHARD (gameID, playerID),

    KEY (race, opponentRace) USING HASH
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

CREATE OR REPLACE FUNCTION compvec_inner(p_minloop BIGINT, p_maxloop BIGINT)
    RETURNS TABLE AS RETURN
        select
            players.gameid, players.playerid, players.race, players.opponentRace, kind,
            ifnull((
                select sum(num) as num
                from buildcomp bc
                where
                    bc.gameid = players.gameid
                    and bc.playerid = players.playerid
                    and bc.kind = kinds.kind
                    and bc.loopid BETWEEN p_minloop AND p_maxloop
                    and num = 1
            ), 0) as num
        from players, uniquekind as kinds, games
        where
            players.gameid = games.gameid
            and games.loops >= p_maxloop;

CREATE OR REPLACE FUNCTION compvec(p_minloop BIGINT, p_maxloop BIGINT)
    RETURNS TABLE AS RETURN
        select 
            gameid, playerid, race, opponentRace,
            json_array_pack(concat("[",group_concat(num order by kind asc separator ','),"]")) as vec
        from compvec_inner(p_minloop, p_maxloop)
        group by gameid, playerid;

CREATE OR REPLACE FUNCTION compvecByRace(p_minloop BIGINT, p_maxloop BIGINT, p_race TEXT, p_opponentRace TEXT)
    RETURNS TABLE AS RETURN
        select 
            gameid, playerid, race, opponentRace,
            json_array_pack(concat("[",group_concat(num order by kind asc separator ','),"]")) as vec
        from compvec_inner(p_minloop, p_maxloop)
        where race = p_race and opponentRace = p_opponentRace
        group by gameid, playerid;

create or replace function comp(p_gameid bigint, p_playerid int, p_minloop BIGINT, p_maxloop bigint)
    returns table as return
        select kind, sum(num) as num
        from buildcomp
        where gameid = p_gameid and playerid = p_playerid
        and loopid between p_minloop and p_maxloop
        and num = 1
        group by kind;

create or replace function compare(p_gameid bigint, p_playerid int, p_loopid bigint, p_gameid2 bigint, p_playerid2 int, p_loopid2 bigint, p_lag BIGINT)
    returns table as return
        select ifnull(a.kind, b.kind) kind, ifnull(a.num, 0) as player1, ifnull(b.num, 0) as player2
        from comp(p_gameid, p_playerid, p_loopid-p_lag, p_loopid) a
        full outer join comp(p_gameid2, p_playerid2, p_loopid2-p_lag, p_loopid2) b
        on a.kind = b.kind
        order by 1 asc;

create or replace function gamesummary(p_gameid bigint)
    returns table as return
        select
            game.gameid, game.loops, game.durationSec, game.mapName,
            player.playerid, player.name, player.race, player.mmr, player.apm, player.result
        from
            games as game,
            players as player
        where
            game.gameid = p_gameid
            and player.gameid = game.gameid;

create or replace function gameHistory(p_gameid BIGINT, p_playerid BIGINT, p_minloop BIGINT, p_maxloop BIGINT)
    RETURNS TABLE AS RETURN
        select * from buildcomp
        where
            gameid = p_gameid
            AND playerid = p_playerid
            AND loopID between p_minloop and p_maxloop
        order by loopID asc;

CREATE OR REPLACE FUNCTION similarGamePoints(
    p_gameid        BIGINT,
    p_playerid      BIGINT,
    p_race          TEXT NOT NULL COLLATE "utf8_bin",
    p_opponentRace  TEXT NOT NULL COLLATE "utf8_bin",
    p_loopid        BIGINT,
    p_lag           BIGINT,
    p_limit         INT
)
RETURNS TABLE AS RETURN
    select
        other.gameid,
        other.playerid,
        other.loopid,
        other.looplag,
        EUCLIDEAN_DISTANCE(other.vec, (
            select vec from compvec(p_loopid - p_lag, p_loopid)
            where gameid = p_gameid and playerid = p_playerid
        )) dist
    from
        compvecs as other
    where
        other.gameid != p_gameid
        and other.race = p_race
        and other.opponentRace = p_opponentRace
    order by
        dist asc,
        ABS(p_loopid-other.loopid) asc,
        other.gameid,
        other.playerid
    limit p_limit;

delimiter //

create or replace procedure prepareCompvecsLag(loopInterval INT, maxloop BIGINT, lag BIGINT) AS
BEGIN
    FOR curloop IN loopInterval .. maxloop BY loopInterval LOOP
        insert into compvecs (gameid, playerid, race, opponentRace, loopid, looplag, vec)
        select gameid, playerid, race, opponentRace, curloop, lag, vec
        from compvec(IFNULL(curloop-lag, 0), curloop);
    END LOOP;
END //

create or replace procedure prepareCompvecs(loopInterval INT) AS
DECLARE
    maxlooptbl QUERY(maxloop BIGINT) = select max(loops) from games;
    maxloop BIGINT = SCALAR(maxlooptbl);
BEGIN
    DELETE FROM compvecs;
    CALL prepareCompvecsLag(loopInterval, maxloop, null);
    CALL prepareCompvecsLag(loopInterval, maxloop, 160); -- ~10 seconds
    CALL prepareCompvecsLag(loopInterval, maxloop, 480); -- ~30 seconds
    CALL prepareCompvecsLag(loopInterval, maxloop, 960); -- ~1 minute
    CALL prepareCompvecsLag(loopInterval, maxloop, 2400); -- ~2.5 minutes
    CALL prepareCompvecsLag(loopInterval, maxloop, 4800); -- ~5 minutes
END //

create or replace procedure prepareUniqueKinds() AS
BEGIN
    DELETE from uniquekind;
    INSERT INTO uniquekind SELECT DISTINCT kind FROM buildcomp;
END //

create or replace procedure postprocess() AS
BEGIN
    CALL prepareUniqueKinds();
    CALL prepareCompvecs(80);
END //

delimiter ;
