import sys

def main():
    s = sys.stdin.readline().strip()
    pair = {')': '(', ']': '[', '}': '{'}
    stack = []
    for c in s:
        if c in pair:
            if not stack or stack.pop() != pair[c]:
                print("false")
                return
        else:
            stack.append(c)
    print("false" if stack else "true")

main()
