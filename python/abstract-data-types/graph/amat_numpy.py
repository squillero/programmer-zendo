# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings

import numpy as np
import numpy.typing as npt
from icecream import ic


class Graph:
    """A weighted directed graph stored with an adjacency matrix (NumPy ndarray)."""

    _data: npt.NDArray[np.float32]

    def __init__(self, num_nodes: int):
        self._data = np.zeros((num_nodes, num_nodes), dtype=np.float32)

    def __len__(self):
        return self._data.shape[0]

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1)
        self._data[n1, n2] = weight

    def del_edge(self, n1: int, n2: int) -> None:
        # Complexity: O(1)
        self._data[n1][n2] = 0.0

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(1)
        return self._data[n1, n2] or None

    def neighbors(self, node: int) -> Iterator[tuple[np.integer, np.float32]]:
        # Complexity: O(|V|) -- vectorized
        row = self._data[node, :]
        yield from zip(row.nonzero()[0], row[row.nonzero()])

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
