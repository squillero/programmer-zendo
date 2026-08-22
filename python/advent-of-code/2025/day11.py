# Advent of Code 2025 | https://adventofcode.com/2025/day/1
# Copyright © 2025 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

import networkx as nx
from icecream import ic

INPUT_FILE_NAME_1 = "day11-test_1.txt"
INPUT_FILE_NAME_2 = "day11-test_2.txt"
# INPUT_FILE_NAME_1 = "day11-input.txt"
# INPUT_FILE_NAME_2 = "day11-input.txt"

YOU = "you"
OUT = "out"
SVR = "svr"
DAC = "dac"
FFT = "fft"


def read_graph(filename):
    G = nx.DiGraph()
    G.graph["exits"] = set()
    with open(filename) as file:
        for line in file:
            device, outputs = line.strip().split(": ")
            for elem in outputs.split():
                # G.add_node(device)
                G.add_edge(device, elem)
    return G


def main():

    # = [Part 1] ==================================================================================
    G = read_graph(INPUT_FILE_NAME_1)
    ic(len(list(nx.all_simple_paths(G, YOU, OUT))))  # d'ho!?

    # = [Part 2] ==================================================================================
    G = read_graph(INPUT_FILE_NAME_2)


if __name__ == "__main__":
    main()
