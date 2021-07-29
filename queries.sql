use sc2;

create or replace function playersByMatchup(p_playerRace TEXT, p_opponentRace TEXT)
    RETURNS TABLE AS RETURN
        select * from players
        where race = p_playerRace and opponentRace = p_opponentRace;

CREATE OR REPLACE FUNCTION compvec(p_loopid BIGINT)
    RETURNS TABLE AS RETURN
        select 
            gameid, playerid, json_array_pack(concat("[",group_concat(num order by kind asc separator ','),"]")) as comp
        from (
            select
                gameid, playerid, kind,
                ifnull((
                    select sum(num)
                    from buildcomp bc
                    where
                        bc.gameid = players.gameid
                        and bc.playerid = players.playerid
                        and bc.kind = kinds.kind
                        and bc.loopid <= p_loopid
                ), 0) as num
            from players, (select distinct kind from buildcomp) as kinds
        )
        group by gameid, playerid;

create or replace function comp(p_gameid bigint, p_playerid int, p_loopid bigint)
    returns table as return
        select kind, sum(num) as num
        from buildcomp
        where gameid = p_gameid and playerid = p_playerid
        and loopid <= p_loopid
        group by kind;

create or replace function compare(p_gameid bigint, p_playerid int, p_gameid2 bigint, p_playerid2 int, p_loopid bigint)
    returns table as return
        select ifnull(a.kind, b.kind) kind, ifnull(a.num, 0) as player1, ifnull(b.num, 0) as player2
        from comp(p_gameid, p_playerid, p_loopid) a
        full outer join comp(p_gameid2, p_playerid2, p_loopid) b
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

select gameid, substring_index(filename, "/", -1) as name, ts, loops, durationSec/60 as mins, matchup
from games;

select * from gamesummary(-5853559066988924600);
select * from comp(-5853559066988924600, 1, 3000);

select
    other.gameid,
    other.playerid,
    EUCLIDEAN_DISTANCE(game.comp, other.comp) dist
from compvec(3000) as game, compvec(3000) as other
where game.gameid = -5853559066988924600 and game.playerid = 1
order by dist asc
limit 10;

select * from compare(-5853559066988924600, 1, 1451672319390495364, 2, 3000)
union all
select "-","-","-"
union all
select * from compare(-5853559066988924600, 1, 1451672319390495364, 2, 3600)
;

select * from comp(-5160808878908170846, 2, 4000);

select * from gamesummary(-9204953945129541784)
union all
select * from gamesummary(-8902385801919203780);