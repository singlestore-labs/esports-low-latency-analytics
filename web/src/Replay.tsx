import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router';
import { useInterval } from 'rooks';

import { ReplayMeta, ReplayStatus } from './models';
import Timeline from './Timeline';

import './Replay.css';

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
        fetch(
            `http://localhost:8000/api/replays/${gameid}/start?startloop=200`
        ).then(refreshStatus);
    };
    const stopReplay = () => {
        fetch(`http://localhost:8000/api/replays/${gameid}/stop`).then(
            refreshStatus
        );
    };

    if (!replay) {
        return <div>Loading...</div>;
    }

    return (
        <div className="Replay">
            <div className="header">
                <div className="mapname">{replay.mapname}</div>
                <div className="player">
                    <b>{replay.p1Name}</b> ({replay.p1Race})
                </div>
                <div className="player">
                    <b>{replay.p2Name}</b> ({replay.p2Race})
                </div>
                <div className="controls">
                    {status.active ? (
                        <div className="button" onClick={stopReplay}>
                            ⏹️
                        </div>
                    ) : (
                        <div className="button" onClick={startReplay}>
                            ▶️️
                        </div>
                    )}
                    <div className="loop">{status.loop}</div>
                </div>
            </div>
            <div className="timelines">
                <Timeline
                    live
                    gameID={replay.gameid}
                    playerID={1}
                    loop={status.loop || 0}
                />
            </div>
        </div>
    );
};

export default Replay;
