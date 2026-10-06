# Shortest Path

A city has `n` junctions, numbered from 1, joined by `m` one-way roads, each
with a length. How long is the shortest drive from junction `s` to junction
`t`?

## Input

The first line has `n m s t`. Each of the next `m` lines has a road `u v w`:
from `u` to `v`, of length `w`.

- `1 ≤ n ≤ 10^5`, `0 ≤ m ≤ 2·10^5`
- `1 ≤ w ≤ 10^9` — the answer may not fit in 32 bits

## Output

The length of the shortest drive, or `-1` if `t` cannot be reached from `s`.

## Example

```
4 5 1 4
1 2 4
1 3 1
3 2 2
2 4 5
3 4 9
```

```
8
```
