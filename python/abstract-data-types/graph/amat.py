# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings

from icecream import ic


class Graph:
    """A weighted directed graph stored with an adjacency matrix (list of lists)."""

    _data: list[list[float | None]]

    def __init__(self, num_nodes: int):
        self._data = [[None] * num_nodes for _ in range(num_nodes)]

    def __len__(self):
        return len(self._data)

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1)
        self._data[n1][n2] = weight

    def del_edge(self, n1: int, n2: int) -> None:
        # Complexity: O(1)
        self._data[n1][n2] = None

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(1)
        return self._data[n1][n2]

    def neighbors(self, node: int) -> Iterator[tuple[int, float]]:
        # Complexity: O(|V|)
        return ((i, v) for i, v in enumerate(self._data[node]) if v is not None)

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
