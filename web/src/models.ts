import * as d3array from 'd3-array';

export type ReplayMeta = {
    gameid: string;
    filename: string;
    mapname: string;
    loops: number;
    p1Name: string;
    p1Race: string;
    p1Result: string;
    p2Name: string;
    p2Race: string;
    p2Result: string;
};

export type ReplayEvent = {
    playerid: number;
    loopid: number;
    kind: string;
    num: number;
};

export type ReplayStats = {
    playerid: number;
    loopid: number;

    foodMade: number;
    foodUsed: number;
    mineralsCollectionRate: number;
    mineralsCurrent: number;
    vespeneCollectionRate: number;
    vespeneCurrent: number;
};

export type EventsStats = {
    events: Array<ReplayEvent>;
    stats: Array<ReplayStats>;
};

export type Timeline = {
    1: EventsStats;
    2: EventsStats;
};

export const loadTimeline = (data: EventsStats): Timeline => {
    let eventsByPlayer = d3array.group(data.events, (e) => e.playerid);
    let statsByPlayer = d3array.group(data.stats, (e) => e.playerid);
    return {
        1: { events: eventsByPlayer.get(1) || [], stats: statsByPlayer.get(1) || [] },
        2: { events: eventsByPlayer.get(2) || [], stats: statsByPlayer.get(2) || [] },
    };
};
