# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

FILENAME = "ali535.tsp.bz2"

import logging
from collections import deque
from itertools import product
from random import seed, randint, random

from icecream import ic
from tqdm.auto import tqdm

from graph.amat import Graph

from graph.read_graph import read_graph
from upq.upq import UPQ


def bfv(G: Graph, node: int) -> list[int]:
    """Breadth-first visit."""

    discovered = {node}
    frontier = deque([node])

    visit = list()
    while frontier:
        node = frontier.popleft()
        visit.append(node)
        for n, _ in G.neighbors(node):
            if n not in discovered:
                discovered.add(n)
                frontier.append(n)
    return visit


def dfv(G: Graph, node: int) -> list[int]:
    """Depth-first visit."""

    discovered = {node}
    frontier = deque([node])

    visit = list()
    while frontier:
        node = frontier.pop()
        visit.append(node)
        for n, _ in G.neighbors(node):
            if n not in discovered:
                discovered.add(n)
                frontier.append(n)
    return visit


def shortest_path(G: Graph, n1: int, n2: int) -> list[int]:
    """Shortest path from n1 to n2 (not correct!)."""

    path: dict = {n1: None}
    dist = [float("inf")] * len(G)

    dist[n1] = 0.0
    frontier = UPQ([(n1, 0.0)])
    while frontier:
        current, cost = frontier.pop()
        if current == n2:
            break
        for n, w in G.neighbors(current):
            if dist[n] > cost + w:
                logging.info(f"spath: relaxing {current}-{n}: {dist[n]:.3f} => {cost + w:.3f}")
                path[n] = current
                dist[n] = cost + w
                frontier.push_update(n, cost + w)
    else:
        return list()

    p = list()
    while current is not None:
        p.append(current)
        current = path[current]
    return p[::-1]


def main():
    logging.basicConfig(level=logging.INFO)

    G = read_graph(Graph, FILENAME)
    # Remove some edge to make the problem slightly harder
    seed(42)
    for n1, n2 in tqdm(product(range(len(G)), repeat=2), total=len(G) ** 2):
        if random() < 0.98:
            G.del_edge(n1, n2)

    logging.info(f"main: Memory usage: {G.size / 1024:,.0f} KiBytes")
    path = shortest_path(G, 42, 99)
    ic(path)
    tot = 0
    for n1, n2 in zip(path, path[1:]):
        tot += G.get_edge(n1, n2)
        ic(n1, n2, G.get_edge(n1, n2))
    ic(tot)


if __name__ == "__main__":
    main()
