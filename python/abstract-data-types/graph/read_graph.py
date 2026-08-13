# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

# This code sucks!

import fileinput
from math import sqrt
from itertools import combinations
import re
from collections import deque
from icecream import ic


def _extract_field(marker: str, full_text: str) -> str:
    match = re.search(rf"{marker}\s*:(.*?)$", full_text, re.IGNORECASE | re.MULTILINE)
    if match:
        return match.group(1).strip().upper()
    else:
        return ""


def _extract_section(marker: str, full_text: str) -> str:
    match = re.search(rf"{marker}(.*?)[A-Z]", full_text, re.IGNORECASE | re.DOTALL)
    assert match, f"D'ho!? Can't find \"{marker.upper()}\"..."
    return match.group(1)


def _extract_numbers(marker: str, full_text: str) -> str:
    match = re.search(
        rf"{marker}\s*([+-]?\d+(?:\.\d+)?(?:\s+[+-]?\d+(?:\.\d+)?)*)",
        full_text,
        re.IGNORECASE | re.DOTALL,
    )
    assert match, f"D'ho!? Can't find \"{marker.upper()}\"..."
    return match.group(1)


def read_graph(G, filename, symmetric=True):
    with fileinput.hook_compressed(filename, mode="r") as data:
        whole_file = data.read()

    name = _extract_field("NAME", whole_file)
    comment = _extract_field("COMMENT", whole_file)
    format = _extract_field("EDGE_WEIGHT_TYPE", whole_file)
    if not format:
        format = _extract_field("EDGE_DATA_FORMAT", whole_file)
    # ic(name)
    # ic(comment)

    n_cities = int(_extract_field("DIMENSION", whole_file))
    graph = G(n_cities)
    if format == "EXPLICIT":
        data = deque(int(d) for d in _extract_numbers("EDGE_WEIGHT_SECTION", whole_file).split())
        if _extract_field("EDGE_WEIGHT_FORMAT", whole_file) == "LOWER_DIAG_ROW":
            for c1 in range(n_cities):
                for c2 in range(c1):
                    w = data.popleft()
                    graph.add_edge(c1, c2, w)
                    if symmetric:
                        graph.add_edge(c2, c1, w)
                assert data.popleft() == 0
        elif _extract_field("EDGE_WEIGHT_FORMAT", whole_file) == "UPPER_ROW":
            for c1 in range(n_cities):
                for c2 in range(1, n_cities - c1):
                    w = data.popleft()
                    graph.add_edge(c1, c2, w)
                    if symmetric:
                        graph.add_edge(c2, c1, w)
    elif format == "ATT" or format == "EUC_2D" or format == "CEIL_2D":
        cities = list()
        for l in _extract_numbers("NODE_COORD_SECTION", whole_file).split("\n"):
            if not l.strip():
                continue
            _, x, y = map(float, l.split())
            cities.append((x, y))
        for c1, c2 in combinations(range(n_cities), r=2):
            d = sqrt((cities[c1][0] - cities[c2][0]) ** 2 + (cities[c1][1] - cities[c2][1]) ** 2)
            graph.add_edge(c1, c2, d)
            if symmetric:
                graph.add_edge(c2, c1, d)
    elif format == "GEO":
        cities = list()
        for l in _extract_numbers("NODE_COORD_SECTION", whole_file).split("\n"):
            if not l.strip():
                continue
            _, x, y = map(float, l.split())
            cities.append((x, y))
        for c1, c2 in combinations(range(n_cities), r=2):
            d = sqrt((cities[c1][0] - cities[c2][0]) ** 2 + (cities[c1][1] - cities[c2][1]) ** 2)
            graph.add_edge(c1, c2, d)
            if symmetric:
                graph.add_edge(c2, c1, d)
    elif format == "EDGE_LIST":
        for l in _extract_numbers("EDGE_DATA_SECTION", whole_file).split("\n"):
            if not l.strip():
                continue
            if l.strip() == "-1":
                break
            n1, n2 = map(int, l.split())
            graph.add_edge(n1 - 1, n2 - 1, 1)
            if symmetric:
                graph.add_edge(n2 - 1, n1 - 1, 1)
    else:
        assert False, f"Can't parse graph: {format}"

    return graph
