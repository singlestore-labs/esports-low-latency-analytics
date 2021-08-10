import React from 'react';

export type WithChildren<T = unknown> = T & { children?: React.ReactNode };

export function* grouped<T, X>(xs: Iterable<T>, key: (_: T) => X): Generator<Array<T>> {
    let cursor: null | X = null;
    let buffer: T[] = [];

    for (const x of xs) {
        const k = key(x);
        if (cursor === null) {
            cursor = k;
        }

        if (cursor !== k) {
            yield buffer;
            cursor = k;
            buffer = [x];
        } else {
            buffer.push(x);
        }
    }
}

export function* mapped<T, X>(xs: Iterable<T>, f: (x: T) => X): Generator<X> {
    for (const x of xs) {
        yield f(x);
    }
}
