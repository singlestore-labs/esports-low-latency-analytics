select count(*) from games;

-- game analysis
select
    race, opponentrace,
    count(*) as matches,
    sum(if(result = "Victory", 1, 0)) / count(*) as win_percent,
    format(avg(mmr), 2) as avg_match_making_rating,
    format(avg(apm), 2) as avg_apm
from players
group by 1,2
order by 1,2;

-- compvecs
select race, opponentRace, loopID, loopLag, vec from compvecs where loopid > 6818;

-- manual checking
select
    other.gameid,
    other.playerid,
    EUCLIDEAN_DISTANCE(game.vec, other.vec) dist
from
    compvec(6818-2400, 6818) as game,
    compvec(6820-2400, 6820) as other
where
    game.gameid = -5280689129783593904 and game.playerid = 2
    and other.gameid in (8282961679202068312,-1494454335525381675)
    and other.playerid = 1
order by dist asc;

select * from gamesummary("-5280689129783593904");
select filename from games where gameid = -5280689129783593904;
select * from comp("-5280689129783593904", 2, 0, 5000);

-- similar games
select distinct gameid, playerid, loopid
from similarGamePoints(-5280689129783593904, 2, 'Zerg', 'Protoss', 7000, 2400, 5);