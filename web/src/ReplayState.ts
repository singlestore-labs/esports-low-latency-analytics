import { LOOPS_PER_SEC } from './const';

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

const controlYT = (func: string, ...args: any[]) => {
    document
        ?.querySelector('iframe')
        ?.contentWindow?.postMessage(JSON.stringify({ event: 'command', func, args }), '*');
};

const YTOFFSET = 2;

export const reduceState = (state: State, action: Action) => {
    switch (action.type) {
        case 'select-player':
            return {
                ...state,
                player: action.player,
            };
        case 'start':
            let loop = state.loop;
            if (loop === 0) {
                loop = LOOPS_PER_SEC * YTOFFSET;
            }

            controlYT('playVideo');
            controlYT('seekTo', loop / LOOPS_PER_SEC - YTOFFSET, true);

            return {
                ...state,
                loop,
                running: true,
            };
        case 'stop':
            controlYT('pauseVideo');
            controlYT('seekTo', 0, true);

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
            controlYT('pauseVideo');

            return {
                ...state,
                running: false,
            };
        case 'skip':
            let newloop = Math.max(0, state.loop + action.amt);
            controlYT('seekTo', newloop / LOOPS_PER_SEC - YTOFFSET, true);

            return {
                ...state,
                loop: Math.min(action.maxLoops, newloop),
            };
        case 'similar':
            return {
                ...state,
                similar: action.similar,
            };
    }
};
