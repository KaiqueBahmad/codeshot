import sys

def main():
    data = sys.stdin.read().split()
    n, amount = int(data[0]), int(data[1])
    coins = list(map(int, data[2:2 + n]))
    inf = amount + 1
    best = [0] + [inf] * amount
    for v in range(1, amount + 1):
        for c in coins:
            if c <= v and best[v - c] + 1 < best[v]:
                best[v] = best[v - c] + 1
    print(best[amount] if best[amount] < inf else -1)

main()
