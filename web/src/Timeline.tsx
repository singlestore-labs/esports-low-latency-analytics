import React, { useState, useEffect, memo } from 'react';
import { useDimensionsRef } from 'rooks';
import classnames from 'classnames';

import { scaleSymlog, scaleLinear, scaleQuantize } from 'd3-scale';
import * as d3array from 'd3-array';

import { Event } from './models';
import { useFetch, formatSeconds } from './util';

type Props = {
    gameID: string;
    player: number;
    loop: number;
    live?: boolean;
};

const timelineHeight = 200;
const loopsPerSecond = 16;
const loopsPerMinute = loopsPerSecond * 60;

const Tickmark = memo(({ tick, left, now }: { tick: number; left: number; now: number }) => (
    <div
        key={tick}
        className={classnames(
            'bg-white px-1 rounded-lg transform -translate-x-1/2 border border-gray-300 absolute top-0 cursor-default select-none',
            {
                'border-indigo-500 bg-indigo-50': tick === 0,
            }
        )}
        style={{ left }}
    >
        {tick === 0 ? formatSeconds(now / loopsPerSecond, false) : formatSeconds(tick / loopsPerSecond, true)}
    </div>
));

const Timeline: React.FC<Props> = (props: Props) => {
    const [events, setEvents] = useState<[Array<Event>, Array<Event>]>([[], []]);
    useFetch(`api/replays/${props.gameID}/timeline`, (allEvents: Array<Event>) => {
        let g = d3array.groups(allEvents, (e) => e.playerid);
        setEvents([g[0][1], g[1][1]]);
    });

    const [ref, dimensions] = useDimensionsRef({ updateOnScroll: false, updateOnResize: true });
    const width = dimensions?.width || 0;
    const height = dimensions?.height || 0;

    const loopRadius = loopsPerMinute * (width > 2000 ? 2 : 1);
    const minLoop = props.loop - loopRadius;
    const maxLoop = props.loop + loopRadius;

    const xAxis = scaleSymlog()
        .constant(10 ** 3)
        .domain([-loopRadius, loopRadius])
        .range([0, width]);

    const yAxis = scaleLinear().domain([10, -10]).range([0, height]);

    const bisector = d3array.bisector((e: Event) => e.loopid);
    const startIndex = bisector.left(events[props.player - 1], minLoop);
    const endIndex = bisector.right(events[props.player - 1], props.live ? props.loop : maxLoop, startIndex);
    const visibleEvents = events[props.player - 1].slice(startIndex, endIndex);

    const binner = d3array
        .bin<Event, number>()
        .value((e: Event) => e.loopid)
        .domain([minLoop, maxLoop])
        .thresholds(50);

    const bins = binner(visibleEvents).map((bin, i, arr) => {
        if (bin.length === 0) {
            return null;
        }

        const zIndex = arr.length - i;
        const left = xAxis((bin.x0 || 0) - props.loop);
        const binWidth = Math.ceil(xAxis((bin.x1 || 0) - props.loop) - left);
        const maxLoop = d3array.max(bin, (e) => e.loopid);

        const kindsBySection = d3array.rollups(
            bin,
            (xs) => d3array.sum(xs, (e) => e.num),
            (e: Event) => (e.num < 0 ? 'bottom' : 'top'),
            (e: Event) => e.kind
        );

        const padding = 15;
        const sectionHeight = timelineHeight / 2 - padding;

        const sections = kindsBySection.map(([section, kinds], i, arr) => {
            const objWidth = kinds.length * binWidth > sectionHeight ? (1 / kinds.length) * sectionHeight : binWidth;
            if (objWidth < 15) {
                return null;
            }

            const kindsRendered = kinds.map(([kind, count]) => (
                <div key={kind} className="relative transform" style={{ maxWidth: objWidth }}>
                    <img src={`http://localhost:8000/api/icon/${kind}`} title={kind} />
                    {Math.abs(count) === 1 ? undefined : (
                        <div
                            className={classnames(
                                'absolute -bottom-0.5 -right-0.5 justify-center px-0.5 py-0.5',
                                'tracking-tighter text-xs font-bold leading-none rounded-full select-none bg-opacity-70 pointer-events-none',
                                {
                                    'text-red-100 bg-red-600': count < 0,
                                    'text-green-100 bg-green-600': count > 0,
                                }
                            )}
                        >
                            {count}
                        </div>
                    )}
                </div>
            ));

            if (section === 'top') {
                return (
                    <div
                        key={section}
                        className="flex flex-col-reverse absolute top-0 rounded-full bg-gradient-to-t from-green-100 to-transparent"
                        style={{ height: sectionHeight }}
                    >
                        {kindsRendered}
                    </div>
                );
            } else if (section === 'bottom') {
                return (
                    <div
                        key={section}
                        className="flex flex-col absolute bottom-0 rounded-full bg-gradient-to-b from-red-100 to-transparent"
                        style={{ height: sectionHeight, marginTop: padding * 2 }}
                    >
                        {kindsRendered}
                    </div>
                );
            }
        });

        return (
            <div
                key={maxLoop}
                className="absolute transition-transform duration-75 ease-linear transform"
                style={{ width: binWidth, height: timelineHeight, ['--tw-translate-x' as any]: `${left}px`, zIndex }}
            >
                {sections}
            </div>
        );
    });

    const ticks = xAxis
        .ticks(width > 2000 ? 20 : 10)
        .filter((tick) => (props.live ? tick <= 0 : true))
        .map((tick) => <Tickmark key={tick} tick={tick} left={xAxis(tick)} now={props.loop} />);

    return (
        <div ref={ref} className="w-full relative" style={{ height: timelineHeight }}>
            <div
                className={classnames('absolute top-1/2 w-1/2 border border-gray-200 border-dashed', {
                    'w-full': !props.live,
                    'w-1/2': props.live,
                })}
            />
            <div
                className={classnames(
                    'absolute left-1/2 h-full bg-gradient-to-b from-transparent via-gray-100 to-transparent w-1 transform -translate-x-1/2',
                    { 'via-gray-200': props.live }
                )}
            />
            <div
                className={classnames('absolute right-0 h-full w-1/2', { hidden: !props.live })}
                style={{
                    backgroundImage: `
                        linear-gradient(to left, white, transparent 70%),
                        linear-gradient(to bottom, white, transparent 70%),
                        linear-gradient(to top, white, transparent 70%),
                        repeating-linear-gradient(70deg, transparent, transparent 10px, rgba(202,236,190,0.3) 20px, rgba(232,236,241,0.1) 20px)
                    `,
                }}
            />
            <div className="absolute left-0 top-1/2">
                <div className="transform -translate-y-1/2 text-xs text-gray-500 relative h-4">{ticks}</div>
            </div>
            {bins}
        </div>
    );
};

export default Timeline;