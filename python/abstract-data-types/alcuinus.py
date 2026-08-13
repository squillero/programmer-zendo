# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

# Yet another river-crossing problem. This is the original formulation by Alcuinus.

from collections import namedtuple
from itertools import chain, combinations, product
from math import inf

from upq.upq import UPQ

from graph.alist import Graph

State = namedtuple("State", ["left", "boat"])

ALL = "ABC"
BOAT_SIZE = 2


# Brothers: A, B, C...
def brother(x: str) -> str:
    return x.upper()


# Sisters: a, b, c...
def sister(x: str) -> str:
    return x.lower()


# No woman can be in the company of another man unless her brother is also present
def is_valid_group(group: set) -> bool:
    for x in ALL:
        if sister(x) in group and brother(x) not in group and set(brother(ALL)) & group:
            return False
    return True


# No need to explicitly save who's on the right bank of the river
def make_right(group: set) -> set:
    return set(set(brother(ALL) + sister(ALL)) - group)


# Check both banks
def is_valid_state(state: State) -> bool:
    left = state.left
    right = make_right(left)
    return is_valid_group(left) and is_valid_group(right)


# I sthere a valid trip from s1 to s2?
def is_trip_possible(s1: State, s2: State) -> bool:
    if s1.boat == s2.boat:
        return False
    if s1.boat == "L":
        if not s2.left < s1.left:
            return False
        boat = s1.left - s2.left
    else:
        if not s1.left < s2.left:
            return False
        boat = s2.left - s1.left

    if not 1 <= len(boat) <= BOAT_SIZE:
        return False

    return is_valid_group(boat)


# Ça va sans dire
def dijkstra(graph, source):
    parent = [None] * len(graph)
    path_length = [inf] * len(graph)
    path_length[source] = 0.0
    frontier = UPQ([(source, 0)])

    while frontier:
        node, dist = frontier.pop()
        for n, d in graph.neighbors(node):
            if path_length[n] > dist + d:
                path_length[n] = dist + d
                parent[n] = node
                frontier.push_update(n, dist + d)

    return path_length, parent


def main():
    valid_states = list()
    for brothers, sisters, boat in product(
        chain.from_iterable(combinations(brother(ALL), r) for r in range(1 + len(ALL))),
        chain.from_iterable(combinations(sister(ALL), r) for r in range(1 + len(ALL))),
        "RL",
    ):
        state = State(left=set(brothers + sisters), boat=boat)
        if is_valid_state(state):
            valid_states.append(state)
    start = valid_states.index(State(left=set(brother(ALL) + sister(ALL)), boat="L"))
    end = valid_states.index(State(left=set(), boat="R"))

    plan = Graph(len(valid_states))
    for i1, i2 in product(range(len(valid_states)), repeat=2):
        if is_trip_possible(valid_states[i1], valid_states[i2]):
            plan.add_edge(i1, i2, 1)

    dist, prev = dijkstra(plan, start)
    path = list()
    if dist[end] != inf:
        print(f"Found a solution of length {dist[end]}:", end="")
        c = end
        while c is not None:
            path.append(c)
            c = prev[c]
        D = "><"
        i = 0
        for s1, s2 in zip(path[::-1], path[-2::-1]):
            boat = valid_states[s1].left ^ valid_states[s2].left
            print(" " + D[i] + "".join(boat) + D[i], end="")
            i = 1 - i
        print()
    else:
        print(f"The problem has no solutions.")


if __name__ == "__main__":
    main()
