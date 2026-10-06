import sys
from bisect import bisect_left

def main():
    data = sys.stdin.read().split()
    n, q = int(data[0]), int(data[1])
    a = list(map(int, data[2:2 + n]))
    out = []
    for x in map(int, data[2 + n:2 + n + q]):
        i = bisect_left(a, x)
        out.append(str(i if i < n and a[i] == x else -1))
    print("\n".join(out))

main()
