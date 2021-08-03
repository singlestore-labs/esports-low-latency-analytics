export type ReplayMeta = {
    gameid: string;
    filename: string;
    mapname: string;
    p1Name: string;
    p1Race: string;
    p1Result: string;
    p2Name: string;
    p2Race: string;
    p2Result: string;
};

export type ReplayStatus = {
    active: boolean;
    loop: number;
};

export type Event = {
    playerid: number;
    loopid: number;
    kind: string;
    num: number;
};
