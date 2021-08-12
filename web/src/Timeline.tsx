import React, { useState, useEffect, memo } from 'react';
import { useDimensionsRef } from 'rooks';
import classnames from 'classnames';

import { scaleSymlog } from 'd3-scale';
import * as d3array from 'd3-array';

import { ReplayEvent } from './models';
import { useFetch, formatSeconds } from './util';

type Props = {
    gameID: string;
    player: 1 | 2;
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
    const events = useFetch(`api/replays/${props.gameID}/timeline`, (allEvents: Array<ReplayEvent>) => {
        let g = d3array.group(allEvents, (e) => e.playerid);
        return {
            1: g.get(1) || [],
            2: g.get(2) || [],
        };
    });

    const [ref, dimensions] = useDimensionsRef();
    const width = dimensions?.width || 0;

    if (!events) {
        return null;
    }

    const loopRadius = loopsPerMinute * (width > 2000 ? 6 : 2);
    const minLoop = props.loop - loopRadius;
    const maxLoop = props.loop + loopRadius;

    const xAxis = scaleSymlog()
        .constant(width > 2000 ? 10 ** 5 : 10 ** 3)
        .domain([-loopRadius, loopRadius])
        .range([0, width]);

    const bisector = d3array.bisector((e: ReplayEvent) => e.loopid);
    const startIndex = bisector.left(events[props.player], minLoop);
    const endIndex = bisector.right(events[props.player], props.live ? props.loop : maxLoop, startIndex);
    const visibleEvents = events[props.player].slice(startIndex, endIndex);

    const binner = d3array
        .bin<ReplayEvent, number>()
        .value((e: ReplayEvent) => e.loopid)
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
            (e: ReplayEvent) => (e.num < 0 ? 'bottom' : 'top'),
            (e: ReplayEvent) => e.kind
        );

        const padding = 15;
        const sectionHeight = timelineHeight / 2 - padding;

        const sections = kindsBySection
            .sort(([section]) => (section === 'top' ? -1 : 1))
            .map(([section, kinds]) => {
                const objWidth =
                    kinds.length * binWidth > sectionHeight ? (1 / kinds.length) * sectionHeight : binWidth;

                const kindsRendered = kinds.map(([kind, count]) => (
                    <div key={kind} className="relative" style={{ maxWidth: objWidth }}>
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

                return (
                    <div
                        key={section}
                        className={classnames('absolute flex rounded-full bg-gradient-to-t', {
                            'top-0 flex-col-reverse from-green-100 to-transparent': section === 'top',
                            'bottom-0 flex-col from-transparent to-red-100': section === 'bottom',
                        })}
                        style={{ height: sectionHeight, marginTop: section === 'bottom' ? padding * 2 : 0 }}
                    >
                        {kindsRendered}
                    </div>
                );
            });

        return (
            <div
                key={maxLoop}
                className="absolute transition-transform duration-75 ease-linear transform flex flex-col items-center"
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
