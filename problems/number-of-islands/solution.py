import sys

def main():
    data = sys.stdin.read().split()
    r, c = int(data[0]), int(data[1])
    grid = [bytearray(row, "ascii") for row in data[2:2 + r]]
    count = 0
    for i in range(r):
        for j in range(c):
            if grid[i][j] != ord("#"):
                continue
            count += 1
            grid[i][j] = ord(".")
            stack = [(i, j)]
            while stack:
                y, x = stack.pop()
                for ny, nx in ((y + 1, x), (y - 1, x), (y, x + 1), (y, x - 1)):
                    if 0 <= ny < r and 0 <= nx < c and grid[ny][nx] == ord("#"):
                        grid[ny][nx] = ord(".")
                        stack.append((ny, nx))
    print(count)

main()
