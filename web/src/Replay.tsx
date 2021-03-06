import { FastForwardIcon, PauseIcon, PlayIcon, RewindIcon, StopIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import * as d3array from 'd3-array';
import React, { memo, useEffect, useReducer } from 'react';
import { useParams } from 'react-router';
import { useIntervalWhen } from 'rooks';
import { LOOPS_PER_SEC } from './const';
import Header from './Header';
import { loadTimeline, ReplayEvent, ReplayMeta } from './models';
import { initialState, reduceState, SimilarGame } from './ReplayState';
import Timeline from './Timeline';
import { formatSeconds, useFetch } from './util';




const HeaderCell = memo(
    ({
        k,
        children,
        ...props
    }: { k: React.ReactNode; children: React.ReactNode } & React.HTMLAttributes<HTMLDivElement>) => (
        <div className="px-2" {...props}>
            <div className="text-xs select-none text-gray-500 uppercase tracking-wider">{k}</div>
            <div className="truncate">{children}</div>
        </div>
    )
);

const Replay: React.FC = () => {
    const { gameid } = useParams<{ gameid: string }>();
    const [state, dispatch] = useReducer(reduceState, null, initialState);

    const replay = useFetch<ReplayMeta>(`api/replays/${gameid}`);
    const timeline = useFetch(`api/replays/${gameid}/timeline`, loadTimeline);

    useIntervalWhen(() => dispatch( { type: 'tick', maxLoops: replay?.loops || Infinity }), 1000 / LOOPS_PER_SEC, true, true);

    const playerEvents = timeline ? timeline[state.player].events : [];
    const bisector = d3array.bisector((e: ReplayEvent) => e.loopid);
    const lastEventIdx = bisector.left(playerEvents, state.loop);

    useEffect(() => {
        let didCancel = false;
        if (!replay) {
            return;
        }

        const loop = state.loop;
        const params = new URLSearchParams({
            playerid: state.player.toString(),
            loop: Math.round(loop).toString(),
            lag: '2400',
            limit: '5',
        });

        (async () => {
            const response = await fetch(`http://localhost:8000/api/replays/${replay.gameid}/similar?${params}`);
            if (response.ok) {
                const data = await response.json();
                if (!didCancel) {
                    dispatch({ type: 'similar', similar: data.map((s: SimilarGame) => ({ ...s, startLoop: loop })) });
                }
            } else {
                console.error(await response.text());
            }
        })();

        return () => {
            didCancel = true;
        };
    }, [replay?.gameid, state.player, lastEventIdx - 1]);

    if (!replay || !timeline) {
        return <h1>Loading...</h1>;
    }

    let yt = null;
    if (replay?.gameid === '-5280689129783593904') {
        yt = (
            <iframe
                width="560"
                height="315"
                src="https://www.youtube.com/embed/7Ry_B3RZQ4M?enablejsapi=1&controls=1&rel=0"
                title="YouTube video player"
                frameBorder="0"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                allowFullScreen
                style={{
                    position: 'fixed',
                    right: '10px',
                    top: '110px',
                    zIndex: 9999,
                }}
            ></iframe>
        );
    }

    return (
        <>
            {yt}
            <Header>
                <div className="flex">
                    <div className="pr-4 mr-4 border-r-2 border-gray-100">
                        <HeaderCell k="map">{replay.mapname}</HeaderCell>
                    </div>
                    <HeaderCell
                        k={replay.p1Race}
                        className={classNames(
                            'cursor-pointer border border-transparent rounded px-2 py-0.5 hover:border-indigo-400',
                            {
                                'bg-indigo-200': state.player === 1,
                            }
                        )}
                        onClick={() => dispatch({ type: 'select-player', player: 1 })}
                    >
                        {replay.p1Name}
                    </HeaderCell>
                    <div className="select-none tracking-wider px-4 self-center text-gray-400">vs</div>
                    <HeaderCell
                        k={replay.p2Race}
                        className={classNames(
                            'cursor-pointer border border-transparent rounded px-2 py-0.5 hover:border-indigo-400',
                            {
                                'bg-indigo-200': state.player === 2,
                            }
                        )}
                        onClick={() => dispatch({ type: 'select-player', player: 2 })}
                    >
                        {replay.p2Name}
                    </HeaderCell>
                </div>
                <div className="flex px-10">
                    <div className="self-center mr-2 text-gray-400 h-6 cursor-pointer hover:text-indigo-400">
                        <RewindIcon
                            className="inline h-6 text-gray-400 hover:text-indigo-400"
                            onClick={() => dispatch({ type: 'skip', amt: -LOOPS_PER_SEC * 15, maxLoops: replay.loops })}
                        />
                        {state.running ? (
                            <PauseIcon
                                className="inline h-6 text-gray-400 hover:text-indigo-400"
                                onClick={() => dispatch({ type: 'pause' })}
                            />
                        ) : (
                            <PlayIcon
                                className="inline h-6 text-gray-400 hover:text-indigo-400"
                                onClick={() => dispatch({ type: 'start' })}
                            />
                        )}
                        <StopIcon
                            className="inline h-6 text-gray-400 hover:text-indigo-400"
                            onClick={() => dispatch({ type: 'stop' })}
                        />
                        <FastForwardIcon
                            className="inline h-6 text-gray-400 hover:text-indigo-400"
                            onClick={() => dispatch({ type: 'skip', amt: LOOPS_PER_SEC * 15, maxLoops: replay.loops })}
                        />
                    </div>
                    <HeaderCell k="game time" title={`${state.loop}`}>
                        <span className="select-none">{formatSeconds(state.loop / LOOPS_PER_SEC)}</span>
                    </HeaderCell>
                </div>
            </Header>
            <div className="grid gap-10">
                <Timeline live gameID={replay.gameid} player={state.player} loop={state.loop || 0} />
                {state.similar.map((game, i) => {
                    let gameloop = game.loop + (state.loop - game.startLoop);
                    if (Math.abs(gameloop - state.loop) <= 80) {
                        gameloop = state.loop;
                    }

                    return <Timeline key={i} gameID={game.gameid} player={game.playerid} loop={gameloop} />;
                })}
            </div>
        </>
    );
};

export default Replay;
