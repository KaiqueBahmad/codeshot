import sys

def count(n, row=0, cols=0, d1=0, d2=0):
    if row == n:
        return 1
    total = 0
    free = ~(cols | d1 | d2) & ((1 << n) - 1)
    while free:
        bit = free & -free
        free ^= bit
        total += count(n, row + 1, cols | bit, ((d1 | bit) << 1) & ((1 << n) - 1), (d2 | bit) >> 1)
    return total

print(count(int(sys.stdin.readline())))
