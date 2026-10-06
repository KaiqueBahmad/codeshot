import sys

def main():
    a = sys.stdin.readline().strip()
    b = sys.stdin.readline().strip()
    prev = [0] * (len(b) + 1)
    for ca in a:
        cur = [0]
        for j, cb in enumerate(b):
            cur.append(prev[j] + 1 if ca == cb else max(prev[j + 1], cur[j]))
        prev = cur
    print(prev[-1])

main()
