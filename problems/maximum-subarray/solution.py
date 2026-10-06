import sys

def main():
    data = sys.stdin.read().split()
    n = int(data[0])
    best = cur = None
    for x in map(int, data[1:1 + n]):
        cur = x if cur is None or cur < 0 else cur + x
        best = cur if best is None or cur > best else best
    print(best)

main()
