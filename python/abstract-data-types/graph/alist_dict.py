# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from collections.abc import Iterator
import warnings


class Graph:
    """A graph stored with an adjacency list of dictionaries { node: weight }."""

    _data: list[dict[int, float]]

    def __init__(self, num_nodes: int):
        self._data = [dict() for _ in range(num_nodes)]

    def __len__(self) -> int:
        return len(self._data)

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1)
        self._data[n1][n2] = weight

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(1)
        return self._data[n1].get(n2)

    def del_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(1)
        if n2 in self._data[n1]:
            del self._data[n1][n2]

    def neighbors(self, node: int) -> Iterator[tuple[int, float]]:
        # Complexity: O(|deg(n)|)
        yield from self._data[node].items()

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
