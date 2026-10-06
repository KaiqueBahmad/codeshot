import sys
import heapq

def main():
    data = sys.stdin.buffer.read().split()
    n, m, s, t = map(int, data[:4])
    adj = [[] for _ in range(n + 1)]
    for i in range(m):
        u, v, w = int(data[4 + 3 * i]), int(data[5 + 3 * i]), int(data[6 + 3 * i])
        adj[u].append((v, w))
    dist = [None] * (n + 1)
    dist[s] = 0
    heap = [(0, s)]
    while heap:
        d, u = heapq.heappop(heap)
        if d > dist[u]:
            continue
        if u == t:
            break
        for v, w in adj[u]:
            nd = d + w
            if dist[v] is None or nd < dist[v]:
                dist[v] = nd
                heapq.heappush(heap, (nd, v))
    print(dist[t] if dist[t] is not None else -1)

main()
