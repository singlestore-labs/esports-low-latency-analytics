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
