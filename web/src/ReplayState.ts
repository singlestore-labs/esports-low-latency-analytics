import { ReplayMeta, ReplayEvent } from './models';

export type SimilarGame = {
    gameid: string;
    playerid: 1 | 2;
    loop: number;
    startLoop: number;
};

export type Action =
    | { type: 'select-player'; player: 1 | 2 }
    | { type: 'start' }
    | { type: 'stop' }
    | { type: 'tick'; maxLoops: number }
    | { type: 'pause' }
    | { type: 'skip'; amt: number; maxLoops: number }
    | { type: 'similar'; similar: SimilarGame[] };

export type State = {
    running: boolean;
    loop: number;
    player: 1 | 2;
    similar: SimilarGame[];
};

export const initialState = (): State => ({
    running: false,
    loop: 0,
    player: 1,
    similar: [],
});

export const reduceState = (state: State, action: Action) => {
    switch (action.type) {
        case 'select-player':
            return {
                ...state,
                player: action.player,
            };
        case 'start':
            return {
                ...state,
                running: true,
            };
        case 'stop':
            return {
                ...state,
                running: false,
                loop: 0,
            };
        case 'tick':
            if (state.running) {
                const nextLoop = Math.min(action.maxLoops, state.loop + 1);
                return { ...state, loop: nextLoop, running: nextLoop !== action.maxLoops };
            }
            return state;
        case 'pause':
            return {
                ...state,
                running: false,
            };
        case 'skip':
            return {
                ...state,
                loop: Math.min(action.maxLoops, Math.max(0, state.loop + action.amt)),
            };
        case 'similar':
            return {
                ...state,
                similar: action.similar,
            };
    }
};
