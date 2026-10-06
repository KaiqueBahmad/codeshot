import sys

def main():
    data = sys.stdin.read().split()
    n = int(data[0])
    iv = sorted((int(data[1 + 2 * i]), int(data[2 + 2 * i])) for i in range(n))
    out = []
    for l, r in iv:
        if out and l <= out[-1][1]:
            out[-1][1] = max(out[-1][1], r)
        else:
            out.append([l, r])
    print(len(out))
    print("\n".join(f"{l} {r}" for l, r in out))

main()
