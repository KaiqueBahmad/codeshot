# Binary Search

Given a sorted list of distinct integers, answer where each queried value is.

## Input

The first line has `n` and `q`. The second line has the `n` integers in
increasing order. The third line has the `q` queried values.

- `1 ≤ n, q ≤ 2·10^5`
- `-10^9 ≤ values ≤ 10^9`

## Output

One line per query: the 0-based position of the value, or `-1` if it is not in
the list.

## Example

```
5 3
-1 0 3 5 9
9 2 -1
```

```
4
-1
0
```
