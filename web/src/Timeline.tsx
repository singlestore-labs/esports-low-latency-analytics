import React, { useState, useEffect } from 'react';
import { useDimensionsRef } from 'rooks';
import classnames from 'classnames';

import { Event } from './models';
import { grouped, mapped } from './util';

type Props = {
    gameID: string;
    playerID: number;
    loop: number;
    live?: boolean;
    loopsPerTick: number;
};

const tickWidth = 5;
const majorTickWidth = tickWidth * 16;
const timelineHeight = 200;

const Timeline: React.FC<Props> = (props: Props) => {
    const [events, setEvents] = useState<Array<Event>>([]);
    const [ref, dimensions] = useDimensionsRef();

    useEffect(() => {
        fetch(`http://localhost:8000/api/replays/${props.gameID}/timeline`)
            .then((response) => response.json())
            .then(setEvents);
    }, [props.gameID]);

    const visibleTicks = Math.floor((dimensions?.width || 0) / tickWidth);
    const visibleLoops = visibleTicks * props.loopsPerTick;
    const minLoop = Math.max(0, props.loop - visibleLoops / 2);
    const maxLoop = props.loop + visibleLoops / 2;

    const shiftwidth = 10;
    let lastLoop = 0;
    let shift = 0;

    // events are already sorted by loopid so the groups will also be sorted by loopid
    const visibleEventsByLoop = grouped(
        events.filter((e) => e.playerid === props.playerID && e.loopid >= minLoop && e.loopid <= maxLoop),
        (event) => event.loopid
    );

    const renderedEvents = mapped(visibleEventsByLoop, (events) => {
        const stateChange = event.num > 0 ? 'created' : 'destroyed';
        let offset = (event.loopid - minLoop) * tickWidth;

        if (lastLoop === event.loopid) {
            offset += shift * shiftwidth;
            shift++;
        } else {
            lastLoop = event.loopid;
            shift = 0;
        }

        return (
            <div className={'event ' + stateChange} key={i} style={{ left: offset }}>
                <img className="icon" src={'http://localhost:8000/api/icon/' + event.kind} title={event.kind} />
            </div>
        );
    });

    const numMajorTicks = dimensions ? dimensions.width / majorTickWidth : 0;
    const pivot = Math.floor(numMajorTicks / 2);
    const timepoints = [];
    for (let i = 0; i <= numMajorTicks; i++) {
        timepoints.push(
            <div
                key={i}
                className={classnames(
                    'bg-white px-1 rounded-lg border border-gray-300 absolute top-0 cursor-default select-none',
                    {
                        'border-indigo-500 bg-indigo-50': i === pivot,
                    }
                )}
                style={{
                    left: i * majorTickWidth,
                }}
            >
                {i === pivot ? 'now' : `${i - pivot}s`}
            </div>
        );
    }

    return (
        <div ref={ref} className="w-full relative" style={{ height: timelineHeight }}>
            <div className="absolute top-1/2 w-full border border-gray-200 border-dashed" />
            <div className="absolute left-0 top-1/2">
                <div className="transform -translate-y-1/2 text-xs text-gray-500 relative h-4">{timepoints}</div>
            </div>
        </div>
    );
};

export default Timeline;
