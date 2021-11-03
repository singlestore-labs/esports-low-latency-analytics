import React, { memo } from 'react';
import { useDimensionsRef } from 'rooks';
import classnames from 'classnames';

import { scaleSymlog } from 'd3-scale';
import * as d3array from 'd3-array';

import { loadTimeline, ReplayEvent, ReplayMeta } from './models';
import { useFetch, formatSeconds, formatSigned } from './util';

import { LOOPS_PER_SEC, LOOPS_PER_MIN } from './const';

type Props = {
    gameID: string;
    player: 1 | 2;
    loop: number;
    live?: boolean;
};

const timelineHeight = 200;

const raceToSupplyIcon = {
    Terran: 'SupplyDepot',
    Zerg: 'Overlord',
    Protoss: 'Pylon',
};

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
        {tick === 0 ? formatSeconds(now / LOOPS_PER_SEC, false) : formatSeconds(tick / LOOPS_PER_SEC, true)}
    </div>
));

const Timeline: React.FC<Props> = (props: Props) => {
    const replay = useFetch<ReplayMeta>(`api/replays/${props.gameID}`);
    const state = useFetch(`api/replays/${props.gameID}/timeline`, loadTimeline);

    const [ref, dimensions] = useDimensionsRef();
    const width = dimensions?.width || 0;

    if (!state) {
        return null;
    }
    const { events, stats } = state[props.player];

    const loopRadius = LOOPS_PER_MIN * (width > 2000 ? 6 : 2);
    const minLoop = props.loop - loopRadius;
    const maxLoop = props.loop + loopRadius;

    const xAxis = scaleSymlog()
        .constant(width > 2000 ? 10 ** 5 : 10 ** 3)
        .domain([-loopRadius, loopRadius])
        .range([0, width]);

    const bisector = d3array.bisector((e: ReplayEvent) => e.loopid);
    const startIndex = bisector.left(events, minLoop);
    const endIndex = bisector.right(events, props.live ? props.loop : maxLoop, startIndex);
    const visibleEvents = events.slice(startIndex, endIndex);

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

    const currentStats = d3array.greatest(stats, (s) => (s.loopid > props.loop ? -1 : s.loopid));
    const supplyIcon = replay ? raceToSupplyIcon[props.player === 1 ? replay.p1Race : replay.p2Race] : 'SupplyDepot';

    return (
        <div className="select-none">
            <div className="w-full p-2 flex items-center bg-gray-50 rounded space-x-4 text-sm mb-2">
                <div className="flex space-x-2 items-center text-sm">
                    <div className={classnames('rounded-lg px-1 py-0.5 ', { 'bg-indigo-100': props.player === 1 })}>
                        {replay?.p1Name}
                    </div>
                    <div className="text-sm text-gray-400">vs</div>
                    <div className={classnames('rounded-lg px-1 py-0.5', { 'bg-indigo-100': props.player === 2 })}>
                        {replay?.p2Name}
                    </div>
                </div>
                <div className="flex space-x-2 items-center rounded bg-gray-200 px-1 py-0.5 text-gray-800 text-sm">
                    <img
                        className="h-5"
                        style={{ filter: 'sepia(100%) hue-rotate(180deg) saturate(3)' }}
                        src="http://localhost:8000/api/icon/MineralField"
                        title={`minerals (rate: ${formatSigned(currentStats?.mineralsCollectionRate || 0)})`}
                    />
                    <div>{currentStats?.mineralsCurrent}</div>
                </div>
                <div className="flex space-x-2 items-center rounded bg-gray-200 px-1 py-0.5 text-gray-800 text-sm">
                    <img
                        className="h-5"
                        style={{ filter: 'sepia(100%) hue-rotate(100deg) saturate(1.3) brightness(1.1)' }}
                        src="http://localhost:8000/api/icon/VespeneGeyser"
                        title={`vespene gas (rate: ${formatSigned(currentStats?.vespeneCollectionRate || 0)})`}
                    />
                    <div>{currentStats?.vespeneCurrent}</div>
                </div>
                <div className="flex space-x-2 items-center rounded bg-gray-200 px-1 py-0.5 text-gray-800 text-sm">
                    <img className="h-5" src={`http://localhost:8000/api/icon/${supplyIcon}`} title="supply" />
                    <div>
                        {(currentStats?.foodUsed || 0) / 4096}/{(currentStats?.foodMade || 0) / 4096}
                    </div>
                </div>
            </div>
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
        </div>
    );
};

export default Timeline;
