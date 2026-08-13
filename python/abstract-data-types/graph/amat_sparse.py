# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from typing import Iterator
import warnings

import numpy as np
from scipy.sparse import SparseEfficiencyWarning, csr_array
from icecream import ic

warnings.simplefilter("ignore", SparseEfficiencyWarning)


class Graph:
    """A weighted directed graph stored with an adjacency matrix (Yale Sparse Matrix)."""

    _data: csr_array
    _num_nodes: int

    def __init__(self, num_nodes: int):
        self._num_nodes = num_nodes
        self._data = csr_array((num_nodes, num_nodes), dtype=np.float32)

    def __len__(self):
        return self._num_nodes

    def add_edge(self, n1: int, n2: int, weight: float) -> None:
        # Complexity: O(1)
        self._data[n1, n2] = weight

    def del_edge(self, n1: int, n2: int) -> None:
        # Complexity: O(1)
        self._data[n1][n2] = 0.0

    def get_edge(self, n1: int, n2: int) -> float | None:
        # Complexity: O(deg(n))
        return self._data[n1, n2] or None

    def neighbors(self, node: int) -> Iterator[tuple[np.integer, np.float32]]:
        # Complexity: O(deg(n))
        start_ptr = self._data.indptr[node]
        end_ptr = self._data.indptr[node + 1]
        yield from zip(self._data.indices[start_ptr:end_ptr], self._data.data[start_ptr:end_ptr])

    @property
    def size(self) -> int:
        try:
            from pympler import asizeof
        except ImportError:
            warnings.warn("pympler is required to compute the size of the graph")
            return 0
        return asizeof.asizeof(self)
