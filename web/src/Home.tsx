import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

import { ReplayMeta } from './models';

import './Home.css';

const Home: React.FC = () => {
    const [replays, setReplays] = useState<Array<ReplayMeta>>([]);

    // TODO: add search and filtering
    useEffect(() => {
        fetch('http://localhost:8000/api/replays')
            .then((response) => response.json())
            .then(setReplays);
    }, []);

    const replayList = replays.map((replay) => {
        return (
            <div className="replay" key={replay.gameid}>
                <div className="filename">
                    <Link to={`/replay/${replay.gameid}`}>
                        {replay.filename}
                    </Link>
                </div>
                <div className="properties">
                    <div className="property-name">Map</div>
                    <div className="property-value">{replay.mapname}</div>

                    <div className="property-name">Winner</div>
                    <div className="property-value">
                        {replay.p1Result === 'Victory'
                            ? replay.p1Name
                            : replay.p2Result === 'Victory'
                            ? replay.p2Name
                            : 'Draw'}
                    </div>

                    <div className="property-name">
                        Player 1 ({replay.p1Race})
                    </div>
                    <div className="property-value">{replay.p1Name}</div>

                    <div className="property-name">
                        Player 2 ({replay.p2Race})
                    </div>
                    <div className="property-value">
                        {replay.p2Name} ({replay.p2Race})
                    </div>
                </div>
            </div>
        );
    });

    return (
        <div className="Home">
            <div className="header"></div>
            <div className="replay-list">{replayList}</div>
        </div>
    );
};

export default Home;
