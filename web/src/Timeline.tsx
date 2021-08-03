import React, { useState, useEffect } from 'react';
import { useDimensionsRef } from 'rooks';

import { Event } from './models';

import './Timeline.css';

type Props = {
    gameID: string;
    playerID: number;
    loop: number;
    live?: boolean;
};

const Timeline: React.FC<Props> = (props: Props) => {
    const [events, setEvents] = useState<Array<Event>>([]);
    const [ref, dimensions] = useDimensionsRef();

    useEffect(() => {
        fetch(`http://localhost:8000/api/replays/${props.gameID}/timeline`)
            .then((response) => response.json())
            .then(setEvents);
    }, [props.gameID]);

    const loopWidth = 5;
    const maxVisibleLoops = Math.floor(
        (dimensions ? dimensions.width : 0) / loopWidth
    );
    const minLoop = Math.max(0, props.loop - maxVisibleLoops / 2);
    const maxLoop = props.loop + maxVisibleLoops / 2;

    const shiftwidth = 10;
    let lastLoop = 0;
    let shift = 0;

    const renderedEvents = events
        .filter((e) => {
            return (
                e.playerid === props.playerID &&
                e.loopid >= minLoop &&
                e.loopid <= maxLoop
            );
        })
        .map((event, i) => {
            const stateChange = event.num > 0 ? 'created' : 'destroyed';
            let offset = (event.loopid - minLoop) * loopWidth;

            if (lastLoop === event.loopid) {
                offset += shift * shiftwidth;
                shift++;
            } else {
                lastLoop = event.loopid;
                shift = 0;
            }

            return (
                <div
                    className={'event ' + stateChange}
                    key={i}
                    style={{ left: offset }}
                >
                    <img
                        className="icon"
                        src={'http://localhost:8000/api/icon/' + event.kind}
                        title={event.kind}
                    />
                </div>
            );
        });

    return (
        <div className="Timeline" ref={ref}>
            <div
                className="cursor"
                style={{ left: dimensions ? dimensions.width / 2 : 0 }}
            />
            {renderedEvents}
        </div>
    );
};

export default Timeline;
