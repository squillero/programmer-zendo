# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings

from icecream import ic


class Graph:
    """A weighted directed graph stored with an adjacency list of tuples (node, weight)."""

    _data: list[list[tuple[int, float]]]
    _num_nodes: int

    def __init__(self, num_nodes: int):
        self._num_nodes = num_nodes
        self._data = [list() for _ in range(num_nodes)]

    def __len__(self):
        return self._num_nodes

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1) -- amortized
        assert self.get_edge(n1, n2) is None
        self._data[n1].append((n2, weight))

    def del_edge(self, n1: int, n2: int) -> None:
        raise NotImplementedError

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(|deg(n)|)
        return next((v for n, v in self._data[n1] if n == n2), None)

    def neighbors(self, node: int) -> Iterator[tuple[int, float]]:
        # Complexity: O(|deg(n)|)
        yield from self._data[node]

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
