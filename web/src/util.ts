import React, { useEffect, useRef } from 'react';
import { useQuery } from 'react-query';

export type WithChildren<T = unknown> = T & { children?: React.ReactNode };

export function* grouped<T, X>(xs: Iterable<T>, key: (_: T) => X): Generator<[X, Array<T>]> {
    let cursor: null | X = null;
    let buffer: T[] = [];

    for (const x of xs) {
        const k = key(x);
        if (cursor === null) {
            cursor = k;
        }

        if (cursor !== k) {
            yield [cursor, buffer];
            cursor = k;
            buffer = [x];
        } else {
            buffer.push(x);
        }
    }
}

export function* mapped<T, X>(xs: Iterable<T>, f: (x: T, i: number) => X): Generator<X> {
    let cursor = 0;
    for (const x of xs) {
        yield f(x, cursor++);
    }
}

export const useFetch = <T, X=unknown>(path: string, transform?: (d: X) => T): T | undefined => {
    const url = `http://localhost:8000/${path}`;

    const { data } = useQuery(
        `${path}-${transform ? 'transformed' : 'raw'}`,
        async () => {
            const response = await fetch(url);
            if (response.ok) {
                const data = await response.json();
                return transform ? transform(data) : data;
            } else {
                throw new Error(`Failed to fetch ${url}: ${await response.text()}`);
            }
        },
        {
            cacheTime: 1000 * 60,
            staleTime: Infinity,
        }
    );

    return data;
};

export const formatSeconds = (seconds: number, signed = false) => {
    const sign = signed ? (seconds < 0 ? '-' : '+') : '';
    const abs = Math.abs(seconds);
    if (abs > 60) {
        const mins = Math.floor(abs / 60);
        const secs = Math.floor(abs % 60);
        return `${sign}${mins}m${secs}s`;
    } else {
        return `${sign}${Math.floor(abs)}s`;
    }
};
