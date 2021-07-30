use sc2;

drop table if exists uniquekind;
create reference table uniquekind as select distinct kind from buildcomp;

CREATE OR REPLACE FUNCTION compvec(p_loopid BIGINT)
    RETURNS TABLE AS RETURN
        select 
            gameid, playerid, race, opponentRace,
            json_array_pack(concat("[",group_concat(num order by kind asc separator ','),"]")) as vec
        from (
            select
                players.gameid, players.playerid, players.race, players.opponentRace, kind,
                ifnull((
                    select sum(num) as num
                    from buildcomp bc
                    where
                        bc.gameid = players.gameid
                        and bc.playerid = players.playerid
                        and bc.kind = kinds.kind
                        and bc.loopid <= p_loopid
                        and num = 1
                ), 0) as num
            from players, uniquekind as kinds, games
            where players.gameid = games.gameid and games.loops >= p_loopid
        )
        group by gameid, playerid;

CREATE OR REPLACE FUNCTION compvecByRace(p_loopid BIGINT, p_playerRace TEXT, p_opponentRace TEXT)
    RETURNS TABLE AS RETURN
        select 
            gameid, playerid,
            json_array_pack(concat("[",group_concat(num order by kind asc separator ','),"]")) as vec
        from (
            select
                players.gameid, players.playerid, kind,
                ifnull((
                    select sum(num) as num
                    from buildcomp bc
                    where
                        bc.gameid = players.gameid
                        and bc.playerid = players.playerid
                        and bc.kind = kinds.kind
                        and bc.loopid <= p_loopid
                        and num = 1
                ), 0) as num
            from players, uniquekind as kinds, games
            where players.gameid = games.gameid and games.loops >= p_loopid
            and players.race = p_playerRace and players.opponentRace = p_opponentRace
        )
        group by gameid, playerid;

create or replace function comp(p_gameid bigint, p_playerid int, p_loopid bigint)
    returns table as return
        select kind, sum(num) as num
        from buildcomp
        where gameid = p_gameid and playerid = p_playerid
        and loopid <= p_loopid
        group by kind;

create or replace function compare(p_gameid bigint, p_playerid int, p_loopid bigint, p_gameid2 bigint, p_playerid2 int, p_loopid2 bigint)
    returns table as return
        select ifnull(a.kind, b.kind) kind, ifnull(a.num, 0) as player1, ifnull(b.num, 0) as player2
        from comp(p_gameid, p_playerid, p_loopid) a
        full outer join comp(p_gameid2, p_playerid2, p_loopid2) b
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

delimiter //
create or replace procedure prepareCompvecs(loopInterval INT, maxloop BIGINT) AS
BEGIN
    delete from compvecs;
    FOR curloop IN 0 .. maxloop BY loopInterval LOOP
        insert into compvecs select gameid, playerid, race, opponentRace, curloop, vec from compvec(curloop);
    END LOOP;
END //
delimiter ;

create or replace function gameHistory(p_gameid BIGINT, p_playerid BIGINT, p_loopStart BIGINT, p_loopEnd BIGINT)
    RETURNS TABLE AS RETURN
        select * from buildcomp
        where
            gameid = p_gameid
            AND playerid = p_playerid
            AND loopID >= p_loopStart and loopID <= p_loopEnd
        order by loopID asc;

create or replace function similarGamePoints(p_gameid BIGINT, p_playerid BIGINT, p_loopid BIGINT, p_limit INT)
    RETURNS TABLE AS RETURN
        select
            other.gameid,
            other.playerid,
            other.loopid,
            EUCLIDEAN_DISTANCE(player.vec, other.vec) dist
        from
            compvec(p_loopid) as player,
            compvecs as other
        where player.gameid = p_gameid and player.playerid = p_playerid
        and other.gameid != player.gameid
        and other.race = player.race and other.opponentRace = player.opponentRace
        order by dist asc
        limit p_limit;