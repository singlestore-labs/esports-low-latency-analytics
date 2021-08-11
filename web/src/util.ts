import React, { useEffect } from 'react';

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

export const useFetch = <T>(path: string, success: (_: T) => unknown): void => {
    const url = `http://localhost:8000/${path}`;

    useEffect(() => {
        let didCancel = false;

        (async () => {
            const response = await fetch(url);
            if (response.ok) {
                const data = await response.json();
                if (!didCancel) {
                    success(data);
                }
            } else {
                console.error(await response.text());
            }
        })();

        return () => {
            didCancel = true;
        };
    }, [path]);
};

export const formatSeconds = (seconds: number, signed = false) => {
    const sign = signed ? seconds < 0 ? '-' : '+' : '';
    const abs = Math.abs(seconds);
    if (abs > 60) {
        const mins = Math.floor(abs / 60);
        const secs = Math.floor(abs % 60);
        return `${sign}${mins}m${secs}s`;
    } else {
        return `${sign}${Math.floor(abs)}s`;
    }
};
