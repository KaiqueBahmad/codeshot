# Merge Intervals

Merge every pair of closed intervals that overlap or touch, until none do.

## Input

The first line has `n`. Each of the next `n` lines has an interval `l r`.

- `1 ≤ n ≤ 10^5`
- `0 ≤ l ≤ r ≤ 10^9`

## Output

The number of merged intervals on the first line, then each one as `l r` on a
line of its own, in increasing order.

## Example

```
4
1 3
8 10
2 6
15 18
```

```
3
1 6
8 10
15 18
```
