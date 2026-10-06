import sys

def main():
    data = sys.stdin.read().split()
    n, target = int(data[0]), int(data[1])
    a = list(map(int, data[2:2 + n]))
    seen = {}
    for j, x in enumerate(a):
        if target - x in seen:
            print(seen[target - x], j)
            return
        seen.setdefault(x, j)

main()
