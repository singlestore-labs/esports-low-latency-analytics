import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router';
import { useInterval } from 'rooks';

import { PlayIcon, StopIcon } from '@heroicons/react/solid';
import { ReplayMeta, ReplayStatus } from './models';
import Header from './Header';
import Timeline from './Timeline';

const Replay: React.FC = () => {
    const { gameid } = useParams<{ gameid: string }>();
    const [replay, setReplay] = useState<ReplayMeta | undefined>(undefined);
    const [status, setStatus] = useState<ReplayStatus>({
        active: false,
        loop: 0,
    });

    useEffect(() => {
        fetch(`http://localhost:8000/api/replays/${gameid}`)
            .then((response) => response.json())
            .then(setReplay);
    }, [gameid]);

    let refreshStatus: (() => Promise<void>) | undefined = undefined;

    const [start, stop] = useInterval(() => {
        if (refreshStatus) {
            refreshStatus();
        }
    }, 50);

    refreshStatus = () =>
        fetch(`http://localhost:8000/api/replays/${gameid}/status`)
            .then((response) => response.json())
            .then((status) => {
                if (status.active) {
                    start();
                } else {
                    stop();
                }
                return status;
            })
            .then(setStatus);

    useEffect(() => {
        if (refreshStatus) {
            refreshStatus();
        }
    }, []);

    const startReplay = () => {
        fetch(`http://localhost:8000/api/replays/${gameid}/start?startloop=200`).then(refreshStatus);
    };
    const stopReplay = () => {
        fetch(`http://localhost:8000/api/replays/${gameid}/stop`).then(refreshStatus);
    };

    if (!replay) {
        return <div>Loading...</div>;
    }

    const HeaderCell = ({ k, children }: { k: React.ReactNode; children: React.ReactNode }) => (
        <div className="p-2">
            <div className="text-xs text-gray-500 uppercase tracking-wider">{k}</div>
            <div className="truncate">{children}</div>
        </div>
    );

    return (
        <>
            <Header>
                <div className="flex">
                    <div className="pr-4 mr-4 border-r-2 border-gray-100">
                        <HeaderCell k="map">{replay.mapname}</HeaderCell>
                    </div>
                    <HeaderCell k={replay.p1Race}>{replay.p1Name}</HeaderCell>
                    <div className="tracking-wider p-4 self-center text-gray-400">vs</div>
                    <HeaderCell k={replay.p2Race}>{replay.p2Name}</HeaderCell>
                </div>
                <div className="flex px-10">
                    <div className="self-center mr-2 text-gray-400 h-6 cursor-pointer hover:text-indigo-400">
                        {status.active ? (
                            <StopIcon className="h-6 text-gray-400 hover:text-indigo-400" onClick={stopReplay} />
                        ) : (
                            <PlayIcon className="h-6 text-gray-400 hover:text-indigo-400" onClick={startReplay} />
                        )}
                    </div>
                    <HeaderCell k="loop">{status.loop || 0}</HeaderCell>
                </div>
            </Header>
            <div className="">
                <Timeline live gameID={replay.gameid} playerID={1} loop={status.loop || 0} />
            </div>
        </>
    );
};

export default Replay;
