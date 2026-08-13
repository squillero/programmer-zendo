# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings

from icecream import ic


class Graph:
    """A weighted directed graph stored as sets of incident arcs."""

    _edges: list[float]
    _data: list[tuple[set[int], set[int]]]
    _num_nodes: int

    def __init__(self, num_nodes: int):
        self._num_nodes = num_nodes
        self._data = [(set(), set()) for _ in range(num_nodes)]
        self._edges = list()

    def __len__(self) -> int:
        return self._num_nodes

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1) -- amortized
        i = len(self._edges)
        self._edges.append(weight)
        self._data[n1][0].add(i)  # out
        self._data[n2][1].add(i)  # in

    def del_edge(self, n1: int, n2: int) -> None:
        raise NotImplementedError

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(?)
        if common := self._data[n1][0] & self._data[n2][1]:
            return self._edges[common.pop()]
        else:
            return None

    def neighbors(self, node: int) -> Iterator[tuple[int, float]]:
        out = self._data[node][0]
        for n, d in enumerate(self._data):
            for e in out & d[1]:
                yield (n, self._edges[e])

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
