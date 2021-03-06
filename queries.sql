use sc2;

/*
-- quickly create a new version of a table
create table temp (
);
insert into temp select * from compvecs;
drop table compvecs;
alter table temp rename to compvecs;
*/

select gameid, substring_index(filename, "/", -1) as name, ts, loops, durationSec/60 as mins, matchup
from games;

select * from gamesummary(-5853559066988924600);
select * from comp(-5853559066988924600, 1, 3000);

select
    other.gameid,
    other.playerid,
    EUCLIDEAN_DISTANCE(game.vec, other.vec) dist
from
    compvec(6818-2400, 6818) as game,
    compvec(7200-2400, 7200) as other
where
    game.gameid = 9202580199611447025 and game.playerid = 2
    and other.gameid in (7244131768046794628, 7244131768046794628, 8437888827081805491)
order by dist asc;

select
    other.gameid,
    other.playerid,
    other.loopid,
    EUCLIDEAN_DISTANCE(player.vec, other.vec) dist
from
    compvec(10080) as player,
    compvecs as other
where player.gameid = -5853559066988924600 and player.playerid = 1
and other.gameid != player.gameid
and other.race = player.race and other.opponentRace = player.opponentRace
order by dist asc
limit 10;

select * from gameHistory(5991778963755882934, 2, 6800, 7500);
select * from similarGamePoints(5991778963755882934, 2, 6800, 5);

select
    histories.*
from
    similarGamePoints(5991778963755882934, 2, 6800, 5) as games,
    buildcomp as histories
where
    games.gameid = histories.gameid
    and games.playerid = histories.playerid
    and histories.loopid between (games.loopid - 1000) and (games.loopid + 1000)
order by histories.gameid, histories.playerid, histories.loopid asc