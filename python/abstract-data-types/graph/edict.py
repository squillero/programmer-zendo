# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings


class Graph:
    """A weighted directed graph stored with a dictionary of edges { (n1, n2): weight }."""

    _data: dict[tuple[int, int], float]
    _num_nodes: int

    def __init__(self, num_nodes: int):
        self._num_nodes = num_nodes
        self._data = dict()

    def __len__(self) -> int:
        return self._num_nodes

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1)
        self._data[(n1, n2)] = weight

    def del_edge(self, n1: int, n2: int) -> None:
        raise NotImplementedError

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(1)
        return self._data.get((n1, n2))

    def neighbors(self, node: int) -> Iterator[tuple[int, float]]:
        # Complexity: O(|E|)
        return ((k[1], w) for k, w in self._data.items() if k[0] == node)

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
