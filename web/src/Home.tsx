import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

import Header from './Header';
import { ReplayMeta } from './models';
import { useFetch } from './util';

const Spinner = (
    <div className="col-span-full justify-self-center my-10">
        <svg
            className="animate-spin h-10 text-gray-400"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
        >
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
        </svg>
    </div>
);

const ReplayCell = ({ k, children }: { k: React.ReactNode; children: React.ReactNode }) => (
    <div className="p-2">
        <div className="text-xs text-gray-500 uppercase tracking-wider">{k}</div>
        <div className="truncate">{children}</div>
    </div>
);

const Replay = ({ replay }: { replay: ReplayMeta }) => {
    let result = 'Draw';
    if (replay.p1Result === 'Victory') {
        result = replay.p1Name;
    } else if (replay.p2Result === 'Victory') {
        result = replay.p2Name;
    }

    return (
        <Link
            to={`/replay/${replay.gameid}`}
            className="grid grid-cols-2 auto-cols-fr rounded-md shadow-sm border border-indigo-50 bg-white hover:shadow-lg hover:border-indigo-400 transition-shadow duration-100 ease-in-out"
        >
            <ReplayCell k="map">{replay.mapname}</ReplayCell>
            <ReplayCell k="winner">{result}</ReplayCell>
            <ReplayCell k={replay.p1Race}>
                <span className="text-sm">{replay.p1Name}</span>
            </ReplayCell>
            <ReplayCell k={replay.p2Race}>
                <span className="text-sm">{replay.p2Name}</span>
            </ReplayCell>
            <div className="col-span-2 p-2 break-all text-xs text-gray-300">{replay.filename}</div>
        </Link>
    );
};

// TODO: add search and filtering

const Home: React.FC = () => {
    const replays = useFetch<Array<ReplayMeta>>('api/replays');

    return (
        <>
            <Header />
            <div className="grid gap-3 2xl:grid-cols-5 xl:grid-cols-4 lg:grid-cols-3 md:grid-cols-2">
                {!replays ? Spinner : null}
                {replays?.map((r) => (
                    <Replay replay={r} key={r.gameid} />
                ))}
            </div>
        </>
    );
};

export default Home;
