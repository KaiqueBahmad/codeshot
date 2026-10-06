import sys

MOD = 10**9 + 7

def fib(n):
    # fast doubling: returns (F(n), F(n+1))
    if n == 0:
        return 0, 1
    a, b = fib(n >> 1)
    c = a * ((2 * b - a) % MOD) % MOD
    d = (a * a + b * b) % MOD
    return (d, (c + d) % MOD) if n & 1 else (c, d)

print(fib(int(sys.stdin.readline()))[0])
