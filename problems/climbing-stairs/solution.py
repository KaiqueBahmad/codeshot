import sys

MOD = 10**9 + 7
n = int(sys.stdin.readline())
a, b, c = 1, 0, 0  # ways to reach i, i-1, i-2
for _ in range(n):
    a, b, c = (a + b + c) % MOD, a, b
print(a)
