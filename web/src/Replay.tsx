import React, { useState, useReducer } from 'react';
import { useParams } from 'react-router';
import { useInterval } from 'rooks';

import { PlayIcon, StopIcon, PauseIcon, FastForwardIcon, RewindIcon } from '@heroicons/react/solid';
import { ReplayMeta } from './models';
import Header from './Header';
import Timeline from './Timeline';
import { useFetch, formatSeconds } from './util';

type State = {
    running: boolean;
    loop: number;
};

type Action =
    | { type: 'start' }
    | { type: 'stop' }
    | { type: 'tick' }
    | { type: 'pause' }
    | { type: 'skip'; amt: number };

const reduceState = (state: State, action: Action) => {
    switch (action.type) {
        case 'start':
            return {
                ...state,
                running: true,
            };
        case 'stop':
            return {
                running: false,
                loop: 0,
            };
        case 'tick':
            return {
                ...state,
                loop: state.loop + 1,
            };
        case 'pause':
            return {
                ...state,
                running: false,
            };
        case 'skip':
            return {
                ...state,
                loop: Math.max(0, state.loop + action.amt),
            };
    }
};

const Replay: React.FC = () => {
    const { gameid } = useParams<{ gameid: string }>();

    const [replay, setReplay] = useState<ReplayMeta | undefined>(undefined);
    useFetch(`api/replays/${gameid}`, setReplay);

    const [state, dispatch] = useReducer(reduceState, {
        running: false,
        loop: 0,
    });

    const [startTick, stopTick] = useInterval(() => dispatch({ type: 'tick' }), 1000 / 16);

    const stop = () => {
        dispatch({ type: 'stop' });
        stopTick();
    };
    const start = () => {
        dispatch({ type: 'start' });
        startTick();
    };
    const pause = () => {
        dispatch({ type: 'pause' });
        stopTick();
    };
    const skipBackward = () => {
        dispatch({ type: 'skip', amt: -16 * 15 });
    };
    const skipForward = () => {
        dispatch({ type: 'skip', amt: 16 * 15 });
    };

    const HeaderCell = ({ k, children }: { k: React.ReactNode; children: React.ReactNode }) => (
        <div className="px-2">
            <div className="text-xs select-none text-gray-500 uppercase tracking-wider">{k}</div>
            <div className="truncate">{children}</div>
        </div>
    );

    if (!replay) {
        return <div>Loading...</div>;
    }

    return (
        <>
            <Header>
                <div className="flex">
                    <div className="pr-4 mr-4 border-r-2 border-gray-100">
                        <HeaderCell k="map">{replay.mapname}</HeaderCell>
                    </div>
                    <HeaderCell k={replay.p1Race}>{replay.p1Name}</HeaderCell>
                    <div className="select-none tracking-wider px-4 self-center text-gray-400">vs</div>
                    <HeaderCell k={replay.p2Race}>{replay.p2Name}</HeaderCell>
                </div>
                <div className="flex px-10">
                    <div className="self-center mr-2 text-gray-400 h-6 cursor-pointer hover:text-indigo-400">
                        <RewindIcon className="inline h-6 text-gray-400 hover:text-indigo-400" onClick={skipBackward} />
                        {state.running ? (
                            <PauseIcon className="inline h-6 text-gray-400 hover:text-indigo-400" onClick={pause} />
                        ) : (
                            <PlayIcon className="inline h-6 text-gray-400 hover:text-indigo-400" onClick={start} />
                        )}
                        <StopIcon className="inline h-6 text-gray-400 hover:text-indigo-400" onClick={stop} />
                        <FastForwardIcon
                            className="inline h-6 text-gray-400 hover:text-indigo-400"
                            onClick={skipForward}
                        />
                    </div>
                    <HeaderCell k="game time">
                        <span className="select-none">{formatSeconds(state.loop / 16)}</span>
                    </HeaderCell>
                </div>
            </Header>
            <div className="">
                <Timeline live gameID={replay.gameid} player={1} loop={state.loop || 0} />
                <Timeline gameID={replay.gameid} player={2} loop={state.loop || 0} />
            </div>
        </>
    );
};

export default Replay;
